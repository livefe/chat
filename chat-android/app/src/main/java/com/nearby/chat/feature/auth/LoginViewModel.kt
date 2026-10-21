package com.nearby.chat.feature.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.format.isValidCnMobile
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
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
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    private val _loggedIn = Channel<Unit>(Channel.CONFLATED)
    val loggedIn: Flow<Unit> = _loggedIn.receiveAsFlow()

    private var countdownJob: Job? = null

    fun onMobileChange(v: String) = _uiState.update { it.copy(mobile = v.filter { c -> c.isDigit() }.take(MOBILE_LENGTH)) }

    fun onCodeChange(v: String) = _uiState.update { it.copy(code = v.filter { c -> c.isDigit() }.take(CODE_LENGTH)) }

    fun toggleAgreed() = _uiState.update { it.copy(agreed = !it.agreed) }

    fun sendCode() {
        val mobile = _uiState.value.mobile
        if (!isValidCnMobile(mobile)) {
            toaster.fail(R.string.auth_invalid_mobile)
            return
        }
        if (_uiState.value.countdown > 0 || _uiState.value.sendingCode) return
        viewModelScope.launch {
            _uiState.update { it.copy(sendingCode = true) }
            runCatching { authRepository.sendLoginCode(mobile) }
                .onSuccess {
                    toaster.success(R.string.auth_code_sent)
                    startCountdown()
                }
                .onFailure { e -> toaster.fail(e, R.string.auth_send_code_failed) }
            _uiState.update { it.copy(sendingCode = false) }
        }
    }

    fun login() {
        val s = _uiState.value
        if (!s.agreed) {
            toaster.show(R.string.login_need_agreement)
            return
        }
        if (!isValidCnMobile(s.mobile)) {
            toaster.fail(R.string.auth_invalid_mobile)
            return
        }
        if (s.code.isBlank()) {
            toaster.fail(R.string.auth_code_hint)
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching { authRepository.login(s.mobile, s.code) }
                .onSuccess {
                    _uiState.update { it.copy(loading = false) }
                    _loggedIn.send(Unit)
                }
                .onFailure { e ->
                    _uiState.update { it.copy(loading = false) }
                    toaster.fail(e, R.string.login_failed)
                }
        }
    }

    private fun startCountdown() {
        countdownJob?.cancel()
        _uiState.update { it.copy(countdown = CODE_COUNTDOWN_SEC) }
        countdownJob = viewModelScope.launch {
            delay(1000)
            for (i in (CODE_COUNTDOWN_SEC - 1) downTo 1) {
                _uiState.update { it.copy(countdown = i) }
                delay(1000)
            }
            _uiState.update { it.copy(countdown = 0) }
        }
    }

    private companion object {
        const val MOBILE_LENGTH = 11
        const val CODE_LENGTH = 6
        const val CODE_COUNTDOWN_SEC = 60
    }
}
