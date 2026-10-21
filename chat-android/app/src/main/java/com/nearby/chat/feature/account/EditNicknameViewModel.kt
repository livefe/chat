package com.nearby.chat.feature.account

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
import com.nearby.chat.data.repo.SessionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class EditNicknameViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(snapshot())
    val uiState: StateFlow<EditNicknameUiState> = _uiState.asStateFlow()

    private val _saved = Channel<Unit>(Channel.CONFLATED)
    val saved: Flow<Unit> = _saved.receiveAsFlow()

    fun load() {
        _uiState.update { snapshot() }
    }

    fun onDraft(v: String) = _uiState.update { it.copy(draft = v.take(NICKNAME_MAX)) }

    fun save() {
        val s = _uiState.value
        val next = s.draft.trim()
        if (next.isBlank()) {
            toaster.show(R.string.nickname_hint)
            return
        }
        if (next == s.original.trim() || s.saving) return
        viewModelScope.launch {
            _uiState.update { it.copy(saving = true) }
            runCatching {
                authRepository.updateProfile(nickname = next, avatarMediaId = null)
                authRepository.detail()
            }.onSuccess {
                _uiState.update { it.copy(saving = false) }
                toaster.success(R.string.common_save_success)
                _saved.send(Unit)
            }.onFailure { e ->
                _uiState.update { it.copy(saving = false) }
                toaster.fail(e, R.string.common_save_failed)
            }
        }
    }

    private fun snapshot(): EditNicknameUiState {
        val name = sessionRepository.cachedUser()?.nickname.orEmpty()
        return EditNicknameUiState(original = name, draft = name)
    }

    private companion object {
        const val NICKNAME_MAX = 32
    }
}
