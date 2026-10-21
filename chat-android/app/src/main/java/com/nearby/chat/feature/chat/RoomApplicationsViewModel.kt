package com.nearby.chat.feature.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.model.RoomApplicationItem
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.repo.ChatRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RoomApplicationsViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(RoomApplicationsUiState())
    val uiState: StateFlow<RoomApplicationsUiState> = _uiState.asStateFlow()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching {
                chatRepository.myRoomApplications(page = 1, pageSize = ChatLimits.CHAT_PAGE_LIMIT)
            }.onSuccess { list ->
                _uiState.update { it.copy(loading = false, list = list) }
            }.onFailure { e ->
                _uiState.update { it.copy(loading = false, list = emptyList()) }
                toaster.fail(e, R.string.common_load_failed)
            }
        }
    }

    fun openableRoomId(item: RoomApplicationItem): Long? {
        if (item.status == STATUS_ENABLED) return item.id
        if (item.statusText.isBlank()) toaster.show(R.string.room_application_pending) else toaster.show(UiText.of(item.statusText))
        return null
    }

    companion object {
        const val STATUS_PENDING = 0
        const val STATUS_ENABLED = 1
        const val STATUS_REJECTED = 2
    }
}
