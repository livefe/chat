package com.nearby.chat.feature.feedback

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.FeedbackRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class FeedbackDetailViewModel @Inject constructor(
    private val feedbackRepository: FeedbackRepository,
    private val toaster: Toaster,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val id = savedStateHandle.get<Long>(ARG_ID) ?: 0L

    private val _uiState = MutableStateFlow(FeedbackDetailUiState())
    val uiState: StateFlow<FeedbackDetailUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    private fun load() {
        if (id <= 0) {
            toaster.fail(R.string.common_invalid_params)
            _uiState.update { it.copy(loading = false) }
            return
        }
        viewModelScope.launch {
            runCatching { feedbackRepository.myDetail(id) }
                .onSuccess { item -> _uiState.update { it.copy(detail = item, loading = false) } }
                .onFailure { e ->
                    toaster.fail(e, R.string.common_load_failed)
                    _uiState.update { it.copy(loading = false) }
                }
        }
    }

    companion object {
        const val ARG_ID = "id"
    }
}
