package com.nearby.chat.feature.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.post.buildCheckInResultPayload
import com.nearby.chat.core.post.encodeCheckInResultPayload
import com.nearby.chat.core.ui.TextResolver
import com.nearby.chat.core.ui.toUiText
import com.nearby.chat.data.repo.PostRepository
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
class PostCheckInScanViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val textResolver: TextResolver,
) : ViewModel() {
    private val _uiState = MutableStateFlow(PostCheckInScanUiState())
    val uiState: StateFlow<PostCheckInScanUiState> = _uiState.asStateFlow()

    private val _results = Channel<CheckInNavResult>(Channel.BUFFERED)
    val results: Flow<CheckInNavResult> = _results.receiveAsFlow()

    private var lastPayload = ""
    private var lastAt = 0L

    fun checkIn(payload: String) {
        val trimmed = payload.trim()
        if (trimmed.isEmpty() || _uiState.value.busy) return
        val now = System.currentTimeMillis()
        if (trimmed == lastPayload && now - lastAt < DUPLICATE_WINDOW_MS) return
        lastPayload = trimmed
        lastAt = now

        viewModelScope.launch {
            _uiState.update { it.copy(busy = true, tip = R.string.scan_checking) }
            runCatching {
                postRepository.signupCheckIn(payload = trimmed)
            }.onSuccess { result ->
                val kind = if (result.alreadyCheckedIn) CheckInNavResult.KIND_ALREADY else CheckInNavResult.KIND_SUCCESS
                val encoded = encodeCheckInResultPayload(buildCheckInResultPayload(kind = kind, result = result))
                _results.send(CheckInNavResult(encoded, kind))
            }.onFailure { e ->
                val message = textResolver.resolve(e.toUiText(R.string.check_in_failed))
                val encoded = encodeCheckInResultPayload(
                    buildCheckInResultPayload(kind = CheckInNavResult.KIND_FAIL, message = message),
                )
                _results.send(CheckInNavResult(encoded, CheckInNavResult.KIND_FAIL))
            }
            _uiState.update { it.copy(busy = false, tip = R.string.scan_tip) }
        }
    }

    companion object {
        private const val DUPLICATE_WINDOW_MS = 1800L
    }
}
