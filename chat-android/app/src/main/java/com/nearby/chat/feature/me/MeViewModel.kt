package com.nearby.chat.feature.me

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.core.model.UserInfo
import com.nearby.chat.data.repo.AuthRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.data.repo.SupportRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val supportRepository: SupportRepository,
    sessionRepository: SessionRepository,
) : ViewModel() {
    val user: StateFlow<UserInfo?> = sessionRepository.userFlow
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), null)
    val supportUnread: StateFlow<Int> = supportRepository.previewUnread
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), 0)

    private val _loggedOut = Channel<Unit>(Channel.CONFLATED)
    val loggedOut: Flow<Unit> = _loggedOut.receiveAsFlow()

    fun refresh() {
        viewModelScope.launch {
            runCatching { authRepository.detail() }
            runCatching { supportRepository.refreshSession() }
        }
    }

    fun logout() {
        viewModelScope.launch {
            runCatching { authRepository.logout() }
            _loggedOut.send(Unit)
        }
    }
}
