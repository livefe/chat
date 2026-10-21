package com.nearby.chat.feature.report

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ReportRepository
import com.nearby.chat.data.repo.UploadRepository
import dagger.hilt.android.lifecycle.HiltViewModel
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
class ReportViewModel @Inject constructor(
    private val reportRepository: ReportRepository,
    private val uploadRepository: UploadRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(ReportUiState())
    val uiState: StateFlow<ReportUiState> = _uiState.asStateFlow()

    private val _submitted = Channel<Unit>(Channel.CONFLATED)
    val submitted: Flow<Unit> = _submitted.receiveAsFlow()

    fun onReason(v: Int) = _uiState.update { it.copy(reason = v) }

    fun onContent(v: String) = _uiState.update { it.copy(content = v.take(MAX_CONTENT)) }

    fun addImages(uris: List<Uri>) {
        val remain = MAX_IMAGES - _uiState.value.images.size
        if (remain <= 0) {
            toaster.fail(R.string.image_limit, MAX_IMAGES)
            return
        }
        val items = uris.take(remain).map { uri ->
            ReportImageItem(
                id = UUID.randomUUID().toString(),
                localUri = uri.toString(),
                uploading = true,
            )
        }
        _uiState.update { it.copy(images = it.images + items) }
        items.forEach { item -> uploadOne(item.id, Uri.parse(item.localUri)) }
    }

    fun removeImage(id: String) {
        _uiState.update { it.copy(images = it.images.filterNot { img -> img.id == id }) }
    }

    fun retryImage(id: String) {
        val item = _uiState.value.images.find { it.id == id } ?: return
        _uiState.update { state ->
            state.copy(
                images = state.images.map {
                    if (it.id == id) it.copy(uploading = true, failed = false, url = "") else it
                },
            )
        }
        uploadOne(id, Uri.parse(item.localUri))
    }

    fun submit(targetType: Int, targetId: Long) {
        val s = _uiState.value
        if (targetId <= 0) {
            toaster.fail(R.string.report_invalid_target)
            return
        }
        if (s.hasUploadingImages) {
            toaster.show(R.string.image_uploading_wait)
            return
        }
        if (s.loading) return
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            val imageIds = s.images.mapNotNull { it.mediaId.takeIf { id -> id > 0L } }
            runCatching {
                reportRepository.submit(
                    targetType = targetType,
                    targetId = targetId,
                    reason = s.reason,
                    content = s.content.trim().ifBlank { null },
                    images = imageIds.ifEmpty { null },
                )
            }.onSuccess {
                _uiState.update { it.copy(loading = false) }
                toaster.success(R.string.report_submitted)
                _submitted.send(Unit)
            }.onFailure { e ->
                _uiState.update { it.copy(loading = false) }
                toaster.fail(e, R.string.common_submit_failed)
            }
        }
    }

    private fun uploadOne(id: String, uri: Uri) {
        viewModelScope.launch {
            runCatching { uploadRepository.uploadImage(uri) }
                .onSuccess { media ->
                    _uiState.update { state ->
                        state.copy(
                            images = state.images.map {
                                if (it.id == id) {
                                    it.copy(url = media.url, mediaId = media.id, uploading = false, failed = false)
                                } else {
                                    it
                                }
                            },
                        )
                    }
                }
                .onFailure { e ->
                    _uiState.update { state ->
                        state.copy(
                            images = state.images.map {
                                if (it.id == id) it.copy(uploading = false, failed = true, url = "", mediaId = 0) else it
                            },
                        )
                    }
                    toaster.fail(e, R.string.image_upload_failed)
                }
        }
    }

    companion object {
        const val MAX_IMAGES = 3
        const val MAX_CONTENT = 500
    }
}
