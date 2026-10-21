package com.nearby.chat.feature.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
import com.nearby.chat.data.repo.LocationRepository
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class UserProfileViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val postRepository: PostRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(UserProfileUiState())
    val uiState: StateFlow<UserProfileUiState> = _uiState.asStateFlow()

    private var userId: Long = 0
    private val pager = OffsetPageTracker(PAGE_SIZE)
    private var lat: Double? = null
    private var lng: Double? = null

    fun load(userId: Long) {
        if (userId <= 0) return
        this.userId = userId
        pager.reset()
        viewModelScope.launch {
            _uiState.update {
                UserProfileUiState(profileLoading = true, postsLoading = true)
            }
            val saved = runCatching { locationRepository.savedLocation() }.getOrNull()
            lat = saved?.lat
            lng = saved?.lng

            runCatching { authRepository.profile(userId) }
                .onSuccess { profile ->
                    _uiState.update {
                        it.copy(profile = profile, profileLoading = false, loadFailed = false)
                    }
                    if (!profile.deactivated && (profile.isSelf || profile.canMessage)) {
                        loadMoreInternal(reset = true)
                    } else {
                        _uiState.update { it.copy(postsLoading = false, hasMore = false) }
                    }
                }
                .onFailure { e ->
                    _uiState.update {
                        it.copy(profileLoading = false, postsLoading = false, loadFailed = true)
                    }
                    toaster.fail(e, R.string.common_load_failed)
                }
        }
    }

    fun refreshPosts() {
        if (userId <= 0) return
        val profile = _uiState.value.profile ?: return
        if (profile.deactivated || (!profile.isSelf && !profile.canMessage)) return
        viewModelScope.launch {
            pager.reset()
            _uiState.update {
                it.copy(
                    hasMore = true,
                    refreshing = true,
                    postsLoading = false,
                )
            }
            loadMoreInternal(reset = true, skipLoadingFlag = true)
            _uiState.update { it.copy(refreshing = false, postsLoading = false) }
        }
    }

    fun loadMorePosts() {
        val s = _uiState.value
        val profile = s.profile
        if (s.postsLoading || !s.hasMore || s.refreshing) return
        if (profile?.deactivated == true || (profile?.isSelf != true && profile?.canMessage != true)) return
        viewModelScope.launch { loadMoreInternal(reset = false) }
    }

    fun blockUser() {
        if (userId <= 0) return
        viewModelScope.launch {
            runCatching { authRepository.blockUser(userId) }
                .onSuccess {
                    toaster.success(R.string.profile_blocked)
                    load(userId)
                }
                .onFailure { e -> toaster.fail(e, R.string.profile_block_failed) }
        }
    }

    fun unblockUser() {
        if (userId <= 0) return
        viewModelScope.launch {
            runCatching { authRepository.unblockUser(userId) }
                .onSuccess {
                    toaster.success(R.string.block_unblocked)
                    load(userId)
                }
                .onFailure { e -> toaster.fail(e, R.string.error_operation_failed) }
        }
    }

    private suspend fun loadMoreInternal(reset: Boolean, skipLoadingFlag: Boolean = false) {
        if (reset) pager.reset()
        if (!skipLoadingFlag) {
            _uiState.update { it.copy(postsLoading = true) }
        }
        runCatching {
            postRepository.userList(
                userId = userId,
                page = pager.page,
                pageSize = pager.pageSize,
                latitude = lat,
                longitude = lng,
            )
        }.onSuccess { (total, list) ->
            _uiState.update { state ->
                val result = pager.apply(state.posts, total, list, replace = reset)
                state.copy(
                    posts = result.list,
                    postsLoading = false,
                    hasMore = result.hasMore,
                )
            }
        }.onFailure { e ->
            _uiState.update { it.copy(postsLoading = false, hasMore = false) }
            toaster.fail(e, R.string.common_load_failed)
        }
    }

    companion object {
        private const val PAGE_SIZE = 20
    }
}
