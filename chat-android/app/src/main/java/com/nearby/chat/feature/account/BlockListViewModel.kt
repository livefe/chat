package com.nearby.chat.feature.account

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.model.BlockListItem
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class BlockListViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(BlockListUiState())
    val uiState: StateFlow<BlockListUiState> = _uiState.asStateFlow()

    private val pager = OffsetPageTracker(PAGE_SIZE)

    init {
        loadMore()
    }

    fun loadMore() {
        val s = _uiState.value
        if (s.loading || !s.hasMore) return
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching {
                authRepository.blockList(pager.page, pager.pageSize)
            }.onSuccess { (total, list) ->
                _uiState.update { state ->
                    val result = pager.apply(state.list, total, list, replace = false)
                    state.copy(
                        list = result.list,
                        loading = false,
                        hasMore = result.hasMore,
                    )
                }
            }.onFailure { e ->
                _uiState.update { it.copy(loading = false, hasMore = false) }
                toaster.fail(e, R.string.common_load_failed)
            }
        }
    }

    fun unblock(item: BlockListItem) {
        if (_uiState.value.unblockingId != 0L) return
        viewModelScope.launch {
            _uiState.update { it.copy(unblockingId = item.userId) }
            runCatching { authRepository.unblockUser(item.userId) }
                .onSuccess {
                    _uiState.update { state ->
                        state.copy(
                            list = state.list.filter { it.userId != item.userId },
                            unblockingId = 0,
                        )
                    }
                    toaster.success(R.string.block_unblocked)
                }
                .onFailure { e ->
                    _uiState.update { it.copy(unblockingId = 0) }
                    toaster.fail(e, R.string.error_operation_failed)
                }
        }
    }

    companion object {
        private const val PAGE_SIZE = 20
    }
}
