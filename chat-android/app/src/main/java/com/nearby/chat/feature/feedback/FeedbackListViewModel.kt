package com.nearby.chat.feature.feedback

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.paging.OffsetPageTracker
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
class FeedbackListViewModel @Inject constructor(
    private val feedbackRepository: FeedbackRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(FeedbackListUiState())
    val uiState: StateFlow<FeedbackListUiState> = _uiState.asStateFlow()

    private val pager = OffsetPageTracker(PAGE_SIZE)

    init {
        loadMore()
    }

    fun loadMore() {
        val s = _uiState.value
        if (s.loading || !s.hasMore) return
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching { feedbackRepository.myList(pager.page, pager.pageSize) }
                .onSuccess { (total, items) ->
                    _uiState.update { state ->
                        val result = pager.apply(state.list, total, items)
                        state.copy(list = result.list, hasMore = result.hasMore, loading = false)
                    }
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.common_load_failed)
                    _uiState.update { it.copy(hasMore = false, loading = false) }
                }
        }
    }

    companion object {
        private const val PAGE_SIZE = 10
    }
}
