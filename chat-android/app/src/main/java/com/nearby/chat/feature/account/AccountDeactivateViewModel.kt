package com.nearby.chat.feature.account

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
import com.nearby.chat.data.repo.SessionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class AccountDeactivateViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val toaster: Toaster,
    sessionRepository: SessionRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(AccountDeactivateUiState())
    val uiState: StateFlow<AccountDeactivateUiState> = _uiState.asStateFlow()

    private val _deactivated = Channel<Unit>(Channel.CONFLATED)
    val deactivated: Flow<Unit> = _deactivated.receiveAsFlow()

    private var countdownJob: Job? = null

    init {
        viewModelScope.launch {
            sessionRepository.userFlow.collect { user ->
                _uiState.update { it.copy(mobile = user?.mobile.orEmpty()) }
            }
        }
    }

    fun onCode(v: String) {
        _uiState.update { it.copy(code = v.filter { ch -> ch.isDigit() }.take(CODE_LENGTH)) }
    }

    fun onAgreed(v: Boolean) {
        _uiState.update { it.copy(agreed = v) }
    }

    fun sendCode() {
        val s = _uiState.value
        if (s.counting > 0 || s.sending) return
        viewModelScope.launch {
            _uiState.update { it.copy(sending = true) }
            runCatching { authRepository.sendDeactivationCode() }
                .onSuccess {
                    toaster.success(R.string.auth_code_sent)
                    startCountdown()
                }
                .onFailure { e -> toaster.fail(e, R.string.auth_send_code_failed) }
            _uiState.update { it.copy(sending = false) }
        }
    }

    fun requestConfirm() {
        if (validate()) _uiState.update { it.copy(confirmVisible = true) }
    }

    fun dismissConfirm() {
        _uiState.update { it.copy(confirmVisible = false) }
    }

    fun submit() {
        _uiState.update { it.copy(confirmVisible = false) }
        if (!validate()) return
        if (_uiState.value.submitting) return
        viewModelScope.launch {
            _uiState.update { it.copy(submitting = true) }
            runCatching { authRepository.deactivate(_uiState.value.code.trim()) }
                .onSuccess {
                    toaster.success(R.string.deactivate_done)
                    _deactivated.send(Unit)
                }
                .onFailure { e -> toaster.fail(e, R.string.deactivate_failed) }
            _uiState.update { it.copy(submitting = false) }
        }
    }

    private fun validate(): Boolean {
        val s = _uiState.value
        if (!s.agreed) {
            toaster.show(R.string.deactivate_need_agreement)
            return false
        }
        if (s.code.isBlank()) {
            toaster.fail(R.string.auth_code_hint)
            return false
        }
        return true
    }

    private fun startCountdown() {
        countdownJob?.cancel()
        countdownJob = viewModelScope.launch {
            _uiState.update { it.copy(counting = CODE_COUNTDOWN_SEC) }
            while (_uiState.value.counting > 0) {
                delay(1000)
                _uiState.update { it.copy(counting = (it.counting - 1).coerceAtLeast(0)) }
            }
        }
    }

    private companion object {
        const val CODE_LENGTH = 6
        const val CODE_COUNTDOWN_SEC = 60
    }
}
