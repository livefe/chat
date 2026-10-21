package com.nearby.chat.feature.post

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.post.PostStatus
import com.nearby.chat.core.ui.TextResolver
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import com.nearby.chat.data.repo.PostRepository
import com.nearby.chat.data.repo.UploadRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID
import javax.inject.Inject

@HiltViewModel
class PostCreateViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val chatRepository: ChatRepository,
    private val uploadRepository: UploadRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
    private val textResolver: TextResolver,
) : ViewModel() {
    private val _uiState = MutableStateFlow(PostCreateUiState())
    val uiState: StateFlow<PostCreateUiState> = _uiState.asStateFlow()

    private val _submitted = Channel<Long>(Channel.BUFFERED)
    val submitted: Flow<Long> = _submitted.receiveAsFlow()

    private val _unavailable = Channel<Unit>(Channel.BUFFERED)
    val unavailable: Flow<Unit> = _unavailable.receiveAsFlow()

    private var nearbyJob: Job? = null
    private var skipNearbyRefresh = false

    fun bindRoom(roomId: Long, fillLocation: Boolean = true) {
        if (roomId <= 0) return
        if (_uiState.value.roomId == roomId && _uiState.value.roomName.isNotBlank() && !fillLocation) {
            if (_uiState.value.locationReady) refreshNearbyRooms()
            return
        }
        _uiState.update { it.copy(roomId = roomId) }
        viewModelScope.launch {
            runCatching { chatRepository.roomDetail(roomId) }
                .onSuccess { room ->
                    val name = room.name.ifBlank { defaultRoomName() }
                    _uiState.update { s ->
                        val hasCoords = room.centerLat != 0.0 || room.centerLng != 0.0
                        val shouldFillLocation = fillLocation && hasCoords
                        s.copy(
                            roomName = name,
                            placeName = if (shouldFillLocation) name else s.placeName,
                            latitude = if (shouldFillLocation) room.centerLat else s.latitude,
                            longitude = if (shouldFillLocation) room.centerLng else s.longitude,
                            locationReady = s.locationReady || shouldFillLocation,
                            locating = if (shouldFillLocation) false else s.locating,
                            locationError = if (shouldFillLocation) null else s.locationError,
                        )
                    }
                    if (_uiState.value.locationReady) refreshNearbyRooms()
                }
                .onFailure { e -> toaster.fail(e, R.string.post_room_load_failed) }
        }
    }

    fun selectAttachRoom(room: NearbyRoomItem) {
        val name = room.name.ifBlank { defaultRoomName() }
        val hasCoords = room.centerLat != 0.0 || room.centerLng != 0.0
        if (hasCoords && _uiState.value.nearbyRooms.isNotEmpty()) skipNearbyRefresh = true
        _uiState.update {
            it.copy(
                roomId = room.id,
                roomName = name,
                placeName = if (hasCoords) name else it.placeName,
                latitude = if (hasCoords) room.centerLat else it.latitude,
                longitude = if (hasCoords) room.centerLng else it.longitude,
                locationReady = it.locationReady || hasCoords,
                locating = if (hasCoords) false else it.locating,
                locationError = if (hasCoords) null else it.locationError,
            )
        }
    }

    fun refreshNearbyRooms() {
        if (skipNearbyRefresh) {
            skipNearbyRefresh = false
            return
        }
        val s = _uiState.value
        if (!s.locationReady) {
            _uiState.update { it.copy(nearbyRooms = emptyList(), nearbyLoading = false) }
            return
        }
        nearbyJob?.cancel()
        nearbyJob = viewModelScope.launch {
            _uiState.update { it.copy(nearbyLoading = true) }
            runCatching {
                chatRepository.syncLocationForNearby(
                    lat = s.latitude,
                    lng = s.longitude,
                    radius = ChatLimits.NEARBY_RADIUS,
                    limit = PostCreateUiState.NEARBY_ATTACH_LIMIT,
                    withMessages = false,
                )
            }
                .onSuccess { result ->
                    _uiState.update {
                        it.copy(
                            nearbyRooms = result.rooms.take(PostCreateUiState.NEARBY_ATTACH_LIMIT),
                            nearbyLoading = false,
                        )
                    }
                }
                .onFailure {
                    _uiState.update { it.copy(nearbyRooms = emptyList(), nearbyLoading = false) }
                }
        }
    }

    fun loadForEdit(postId: Long) {
        if (postId <= 0) return
        if (_uiState.value.editPostId == postId && !_uiState.value.pageLoading) return
        viewModelScope.launch {
            _uiState.update {
                it.copy(editPostId = postId, pageLoading = true, loadFailed = false, inForm = true)
            }
            runCatching { postRepository.detail(postId) }
                .onSuccess { d ->
                    val blocked = when (d.status) {
                        PostStatus.PUBLISHED -> R.string.post_edit_published
                        PostStatus.OFFLINE -> R.string.post_edit_offline
                        else -> null
                    }
                    if (blocked != null) {
                        toaster.fail(blocked)
                        markUnavailable()
                        return@onSuccess
                    }
                    _uiState.update { PostCreateForm.fromPost(it, d) }
                    if (d.roomId > 0) {
                        bindRoom(d.roomId, fillLocation = false)
                    } else if (_uiState.value.locationReady) {
                        refreshNearbyRooms()
                    }
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.common_load_failed)
                    markUnavailable()
                }
        }
    }

    private suspend fun markUnavailable() {
        _uiState.update { it.copy(pageLoading = false, loadFailed = true) }
        _unavailable.send(Unit)
    }

    fun onTitle(v: String) = _uiState.update { it.copy(title = v) }
    fun onContent(v: String) = _uiState.update { it.copy(content = v) }
    fun onPinned(v: Boolean) = _uiState.update { it.copy(pinned = v) }

    fun onCategory(v: String) {
        _uiState.update { it.copy(category = v, pinned = false, inForm = true) }
        ensureLocation()
    }

    fun backToTypePick() {
        if (_uiState.value.isEdit) return
        _uiState.update { it.copy(inForm = false) }
    }

    fun onEnableSchedule(v: Boolean) = _uiState.update { it.copy(enableSchedule = v) }
    fun onStartAt(v: Long) = _uiState.update { it.copy(startAt = v) }
    fun onEndAt(v: Long) = _uiState.update { it.copy(endAt = v) }

    fun onEnableFee(v: Boolean) = _uiState.update { it.copy(enableFee = v) }
    fun onFeeKind(v: FeeKind) = _uiState.update { it.copy(feeKind = v) }
    fun onFeeAmount(v: String) = _uiState.update { it.copy(feeAmount = PostCreateForm.filterNumber(v)) }
    fun onFeeDescription(v: String) = _uiState.update { it.copy(feeDescription = v) }

    fun onNeedSignup(v: Boolean) = _uiState.update {
        it.copy(
            needSignup = v,
            needCheckIn = if (v) it.needCheckIn else false,
            enableSchedule = if (v) true else it.enableSchedule,
        )
    }

    fun onNeedCheckIn(v: Boolean) = _uiState.update { it.copy(needCheckIn = v) }
    fun onLimitParticipants(v: Boolean) = _uiState.update { it.copy(limitParticipants = v) }
    fun onMaxParticipants(v: String) = _uiState.update { it.copy(maxParticipants = PostCreateForm.filterDigits(v)) }
    fun onAllowWaitlist(v: Boolean) = _uiState.update { it.copy(allowWaitlist = v) }
    fun onSignupDeadlineAt(v: Long) = _uiState.update { it.copy(signupDeadlineAt = v) }

    fun onEnableIdlePrice(v: Boolean) = _uiState.update { it.copy(enableIdlePrice = v) }
    fun onIdlePrice(v: String) = _uiState.update { it.copy(idlePrice = PostCreateForm.filterNumber(v)) }
    fun onIdleNegotiable(v: Boolean) = _uiState.update { it.copy(idleNegotiable = v) }

    fun onHelpReward(v: Boolean) = _uiState.update { it.copy(helpReward = v) }
    fun onHelpRewardAmount(v: String) = _uiState.update { it.copy(helpRewardAmount = PostCreateForm.filterNumber(v)) }
    fun onHelpRewardDesc(v: String) = _uiState.update { it.copy(helpRewardDesc = v) }

    fun addImages(uris: List<Uri>) {
        if (uris.isEmpty()) return
        val remain = PostCreateUiState.MAX_IMAGES - _uiState.value.images.size
        if (remain <= 0) {
            toaster.fail(R.string.post_image_limit, PostCreateUiState.MAX_IMAGES)
            return
        }
        val items = uris.take(remain).map { uri ->
            PostImageItem(
                id = UUID.randomUUID().toString(),
                localUri = uri.toString(),
                uploading = true,
            )
        }
        _uiState.update { it.copy(images = it.images + items) }
        items.forEach { item -> uploadOne(item.id, Uri.parse(item.localUri)) }
    }

    fun removeImage(id: String) {
        _uiState.update { s -> s.copy(images = s.images.filterNot { it.id == id }) }
    }

    fun retryImage(id: String) {
        val item = _uiState.value.images.find { it.id == id } ?: return
        updateImage(id) { it.copy(uploading = true, failed = false, url = "") }
        uploadOne(id, Uri.parse(item.localUri))
    }

    private fun uploadOne(id: String, uri: Uri) {
        viewModelScope.launch {
            runCatching { uploadRepository.uploadImage(uri) }
                .onSuccess { media ->
                    updateImage(id) {
                        it.copy(url = media.url, mediaId = media.id, uploading = false, failed = false)
                    }
                }
                .onFailure { e ->
                    updateImage(id) { it.copy(uploading = false, failed = true, url = "", mediaId = 0) }
                    toaster.fail(e, R.string.post_image_upload_failed)
                }
        }
    }

    private fun updateImage(id: String, transform: (PostImageItem) -> PostImageItem) {
        _uiState.update { s -> s.copy(images = s.images.map { if (it.id == id) transform(it) else it }) }
    }

    fun applyLocation(lat: Double, lng: Double, placeName: String) {
        _uiState.update {
            it.copy(
                latitude = lat,
                longitude = lng,
                placeName = PlaceNames.realOrEmpty(placeName),
                locationReady = true,
                locating = false,
                locationError = null,
                roomId = 0,
                roomName = "",
                pinned = false,
            )
        }
        refreshNearbyRooms()
    }

    fun ensureLocation() {
        val s = _uiState.value
        if (s.locationReady || s.locating) return
        viewModelScope.launch {
            _uiState.update { it.copy(locating = true, locationError = null) }

            val saved = runCatching { locationRepository.savedLocation() }.getOrNull()
            if (applyCoords(saved?.lat, saved?.lng, saved?.placeName)) return@launch

            val remembered = chatRepository.lastLocationOrNull()
            if (remembered != null && applyCoords(remembered.first, remembered.second, null)) return@launch

            val loc = runCatching { locationRepository.currentLocation() }.getOrNull()
            if (loc != null && applyCoords(loc.lat, loc.lng, null)) return@launch

            val savedAgain = runCatching { locationRepository.savedLocation() }.getOrNull()
            if (applyCoords(savedAgain?.lat, savedAgain?.lng, savedAgain?.placeName)) return@launch

            _uiState.update {
                it.copy(locating = false, locationError = UiText.of(R.string.post_locate_failed))
            }
        }
    }

    private fun applyCoords(lat: Double?, lng: Double?, placeName: String?): Boolean {
        if (lat == null || lng == null) return false
        if (lat == 0.0 && lng == 0.0) return false
        _uiState.update {
            it.copy(
                latitude = lat,
                longitude = lng,
                placeName = PlaceNames.realOrEmpty(placeName.orEmpty()),
                locationReady = true,
                locating = false,
                locationError = null,
            )
        }
        refreshNearbyRooms()
        return true
    }

    fun submit() {
        val s = _uiState.value
        PostCreateForm.validate(s)?.let {
            toaster.fail(it)
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            val draft = PostCreateForm.toDraft(_uiState.value)
            if (s.isEdit) {
                runCatching { postRepository.update(draft) }
                    .onSuccess {
                        _uiState.update { it.copy(loading = false) }
                        toaster.success(R.string.post_update_submitted)
                        _submitted.send(s.editPostId)
                    }
                    .onFailure { e ->
                        _uiState.update { it.copy(loading = false) }
                        toaster.fail(e, R.string.common_save_failed)
                    }
            } else {
                runCatching { postRepository.create(draft) }
                    .onSuccess { id ->
                        _uiState.update { it.copy(loading = false) }
                        toaster.success(R.string.post_create_submitted)
                        _submitted.send(id)
                    }
                    .onFailure { e ->
                        _uiState.update { it.copy(loading = false) }
                        toaster.fail(e, R.string.post_publish_failed)
                    }
            }
        }
    }

    private fun defaultRoomName(): String = textResolver.resolve(UiText.of(R.string.room_default_name))
}
