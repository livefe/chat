package com.nearby.chat.feature.account

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.data.repo.UploadRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class EditProfileViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val sessionRepository: SessionRepository,
    private val uploadRepository: UploadRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(snapshot())
    val uiState: StateFlow<EditProfileUiState> = _uiState.asStateFlow()

    fun load() {
        _uiState.update { current ->
            val next = snapshot()
            if (current.uploadingAvatar && current.localAvatarUri.isNotBlank()) {
                next.copy(
                    localAvatarUri = current.localAvatarUri,
                    uploadingAvatar = true,
                    saving = current.saving,
                )
            } else {
                next.copy(saving = current.saving)
            }
        }
    }

    fun pickAvatar(uri: Uri) {
        if (_uiState.value.uploadingAvatar || _uiState.value.saving) return
        _uiState.update {
            it.copy(localAvatarUri = uri.toString(), uploadingAvatar = true)
        }
        viewModelScope.launch {
            runCatching { uploadRepository.uploadImage(uri) }
                .onSuccess { media ->
                    _uiState.update {
                        it.copy(avatarUrl = media.url, uploadingAvatar = false)
                    }
                    persist(avatarMediaId = media.id)
                }
                .onFailure { e ->
                    _uiState.update { it.copy(uploadingAvatar = false, localAvatarUri = "") }
                    toaster.fail(e, R.string.avatar_upload_failed)
                }
        }
    }

    private fun snapshot(): EditProfileUiState {
        val user = sessionRepository.cachedUser()
        return EditProfileUiState(
            nickname = user?.nickname.orEmpty(),
            mobile = user?.mobile.orEmpty(),
            avatarUrl = user?.avatarUrl.orEmpty(),
        )
    }

    private fun persist(avatarMediaId: Long) {
        if (_uiState.value.saving) return
        viewModelScope.launch {
            _uiState.update { it.copy(saving = true) }
            runCatching {
                authRepository.updateProfile(
                    nickname = null,
                    avatarMediaId = avatarMediaId,
                )
            }.onSuccess {
                _uiState.update { snapshot().copy(saving = false) }
                toaster.success(R.string.common_save_success)
            }.onFailure { e ->
                _uiState.update { it.copy(saving = false) }
                toaster.fail(e, R.string.common_save_failed)
            }
        }
    }
}
