package com.nearby.chat.feature.account

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.format.isValidCnMobile
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
class EditMobileViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(
        EditMobileUiState(currentMobile = sessionRepository.cachedUser()?.mobile.orEmpty()),
    )
    val uiState: StateFlow<EditMobileUiState> = _uiState.asStateFlow()

    private val _saved = Channel<Unit>(Channel.CONFLATED)
    val saved: Flow<Unit> = _saved.receiveAsFlow()

    private var oldCountdownJob: Job? = null
    private var newCountdownJob: Job? = null

    fun load() {
        _uiState.update { it.copy(currentMobile = sessionRepository.cachedUser()?.mobile.orEmpty()) }
    }

    fun onOldCode(v: String) {
        _uiState.update { it.copy(oldCode = v.filter { ch -> ch.isDigit() }.take(CODE_LENGTH)) }
    }

    fun onNewMobile(v: String) {
        _uiState.update { it.copy(newMobile = v.filter { ch -> ch.isDigit() }.take(MOBILE_LENGTH)) }
    }

    fun onNewCode(v: String) {
        _uiState.update { it.copy(newCode = v.filter { ch -> ch.isDigit() }.take(CODE_LENGTH)) }
    }

    fun onBack(): Boolean {
        if (_uiState.value.step == EditMobileUiState.STEP_BIND_NEW) {
            _uiState.update { it.copy(step = EditMobileUiState.STEP_VERIFY_OLD) }
            return true
        }
        return false
    }

    fun goNext() {
        val code = _uiState.value.oldCode.trim()
        if (code.isBlank()) {
            toaster.show(R.string.auth_code_hint)
            return
        }
        if (_uiState.value.verifying) return
        viewModelScope.launch {
            _uiState.update { it.copy(verifying = true) }
            runCatching { authRepository.verifyChangeMobileOldCode(code) }
                .onSuccess { ticket ->
                    if (ticket.isBlank()) {
                        toaster.show(R.string.mobile_verify_failed_retry)
                    } else {
                        _uiState.update { it.copy(ticket = ticket, step = EditMobileUiState.STEP_BIND_NEW) }
                    }
                }
                .onFailure { e -> toaster.fail(e, R.string.mobile_verify_failed) }
            _uiState.update { it.copy(verifying = false) }
        }
    }

    fun sendOldCode() {
        val s = _uiState.value
        if (s.oldCounting > 0 || s.sendingOld) return
        viewModelScope.launch {
            _uiState.update { it.copy(sendingOld = true) }
            runCatching { authRepository.sendChangeMobileOldCode() }
                .onSuccess {
                    toaster.success(R.string.auth_code_sent)
                    startOldCountdown()
                }
                .onFailure { e -> toaster.fail(e, R.string.auth_send_code_failed) }
            _uiState.update { it.copy(sendingOld = false) }
        }
    }

    fun sendNewCode() {
        val s = _uiState.value
        if (s.newCounting > 0 || s.sendingNew) return
        if (!validateNewMobile(s)) return
        viewModelScope.launch {
            _uiState.update { it.copy(sendingNew = true) }
            runCatching { authRepository.sendChangeMobileNewCode(s.newMobile, s.ticket) }
                .onSuccess {
                    toaster.success(R.string.auth_code_sent)
                    startNewCountdown()
                }
                .onFailure { e -> toaster.fail(e, R.string.auth_send_code_failed) }
            _uiState.update { it.copy(sendingNew = false) }
        }
    }

    fun submit() {
        val s = _uiState.value
        if (s.submitting) return
        if (!validateNewMobile(s)) return
        if (s.newCode.isBlank()) {
            toaster.show(R.string.mobile_new_code_hint)
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(submitting = true) }
            runCatching {
                authRepository.changeMobile(
                    ticket = s.ticket,
                    newMobile = s.newMobile.trim(),
                    newCode = s.newCode.trim(),
                )
            }.onSuccess {
                _uiState.update { it.copy(submitting = false) }
                toaster.success(R.string.mobile_changed)
                _saved.send(Unit)
            }.onFailure { e ->
                _uiState.update { it.copy(submitting = false) }
                toaster.fail(e, R.string.mobile_change_failed)
            }
        }
    }

    private fun validateNewMobile(s: EditMobileUiState): Boolean {
        if (s.ticket.isBlank()) {
            toaster.show(R.string.mobile_verify_current_first)
            _uiState.update { it.copy(step = EditMobileUiState.STEP_VERIFY_OLD) }
            return false
        }
        if (!isValidCnMobile(s.newMobile)) {
            toaster.show(R.string.mobile_invalid_new)
            return false
        }
        if (s.newMobile == s.currentMobile) {
            toaster.show(R.string.mobile_same_as_current)
            return false
        }
        return true
    }

    private fun startOldCountdown() {
        oldCountdownJob?.cancel()
        oldCountdownJob = viewModelScope.launch {
            _uiState.update { it.copy(oldCounting = CODE_COUNTDOWN_SEC) }
            while (_uiState.value.oldCounting > 0) {
                delay(1000)
                _uiState.update { it.copy(oldCounting = (it.oldCounting - 1).coerceAtLeast(0)) }
            }
        }
    }

    private fun startNewCountdown() {
        newCountdownJob?.cancel()
        newCountdownJob = viewModelScope.launch {
            _uiState.update { it.copy(newCounting = CODE_COUNTDOWN_SEC) }
            while (_uiState.value.newCounting > 0) {
                delay(1000)
                _uiState.update { it.copy(newCounting = (it.newCounting - 1).coerceAtLeast(0)) }
            }
        }
    }

    companion object {
        const val MOBILE_LENGTH = 11
        const val CODE_LENGTH = 6
        const val CODE_MIN_LENGTH = 4
        private const val CODE_COUNTDOWN_SEC = 60
    }
}
