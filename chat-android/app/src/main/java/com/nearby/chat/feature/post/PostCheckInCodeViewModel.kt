package com.nearby.chat.feature.post

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.media.QrCodeEncoder
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import javax.inject.Inject

@HiltViewModel
class PostCheckInCodeViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val toaster: Toaster,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val postId: Long = savedStateHandle.get<Long>(ARG_POST_ID) ?: 0L

    private val _uiState = MutableStateFlow(PostCheckInCodeUiState())
    val uiState: StateFlow<PostCheckInCodeUiState> = _uiState.asStateFlow()

    init {
        if (postId > 0) load()
    }

    fun load() {
        if (postId <= 0) return
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching { postRepository.signupCheckInCode(postId) }
                .onSuccess { data ->
                    val qr = if (data.payload.isNotBlank()) {
                        withContext(Dispatchers.Default) { QrCodeEncoder.encode(data.payload, QR_SIZE_PX) }
                    } else {
                        null
                    }
                    _uiState.update {
                        it.copy(
                            loading = false,
                            postTitle = data.postTitle,
                            checkedIn = data.checkedIn,
                            checkedInAtText = data.checkedInAtText,
                            userNickname = data.userNickname,
                            signupAt = data.signupAt,
                            qrBitmap = qr,
                        )
                    }
                    if (qr == null && data.payload.isNotBlank()) {
                        toaster.fail(R.string.qr_generate_failed)
                    }
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.common_load_failed)
                    _uiState.update { it.copy(loading = false) }
                }
        }
    }

    companion object {
        const val ARG_POST_ID = "postId"
        private const val QR_SIZE_PX = 560
    }
}
