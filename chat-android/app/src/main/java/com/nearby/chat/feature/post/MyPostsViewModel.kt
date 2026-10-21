package com.nearby.chat.feature.post

import androidx.annotation.StringRes
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MyPostsViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val toaster: Toaster,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val mode = when (savedStateHandle.get<String>(ARG_MODE)) {
        MODE_JOINED -> MyPostsMode.Joined
        else -> MyPostsMode.Published
    }

    private val _uiState = MutableStateFlow(MyPostsUiState(mode = mode))
    val uiState: StateFlow<MyPostsUiState> = _uiState.asStateFlow()

    private val pager = OffsetPageTracker(PAGE_SIZE)
    private var loadJob: Job? = null

    init {
        reload(showPullIndicator = false)
    }

    fun refresh() {
        reload(showPullIndicator = true)
    }

    fun loadMore() {
        val s = _uiState.value
        if (s.loading || !s.hasMore || s.refreshing) return
        loadJob = viewModelScope.launch { loadMoreInternal() }
    }

    fun withdraw(post: PostItem) = act(post, R.string.post_withdrawn, R.string.post_withdraw_failed) { postRepository.withdraw(it) }

    fun deletePublished(post: PostItem) = act(post, R.string.common_deleted, R.string.common_delete_failed) { postRepository.deleteMine(it) }

    fun deleteJoined(post: PostItem) = act(post, R.string.post_removed, R.string.post_remove_failed) { postRepository.removeJoined(it) }

    private fun act(
        post: PostItem,
        @StringRes success: Int,
        @StringRes failure: Int,
        block: suspend (Long) -> Unit,
    ) {
        if (_uiState.value.actingId != 0L) return
        viewModelScope.launch {
            _uiState.update { it.copy(actingId = post.id) }
            runCatching { block(post.id) }
                .onSuccess {
                    toaster.success(success)
                    reload(showPullIndicator = false)
                }
                .onFailure { e -> toaster.fail(e, failure) }
            _uiState.update { it.copy(actingId = 0) }
        }
    }

    private fun reload(showPullIndicator: Boolean) {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            pager.reset()
            _uiState.update {
                it.copy(
                    hasMore = true,
                    refreshing = showPullIndicator,
                    loading = !showPullIndicator,
                )
            }
            loadMoreInternal(skipLoadingFlag = showPullIndicator)
            _uiState.update { it.copy(refreshing = false, loading = false) }
        }
    }

    private suspend fun loadMoreInternal(skipLoadingFlag: Boolean = false) {
        if (!skipLoadingFlag) {
            _uiState.update { it.copy(loading = true) }
        }
        runCatching {
            when (mode) {
                MyPostsMode.Joined -> postRepository.myJoinedList(page = pager.page, pageSize = pager.pageSize)
                MyPostsMode.Published -> postRepository.myList(page = pager.page, pageSize = pager.pageSize)
            }
        }.onSuccess { (total, items) ->
            _uiState.update { state ->
                val result = pager.apply(state.list, total, items)
                state.copy(list = result.list, hasMore = result.hasMore, loading = false)
            }
        }.onFailure { e ->
            toaster.fail(e, R.string.common_load_failed)
            _uiState.update { it.copy(hasMore = false, loading = false) }
        }
    }

    companion object {
        const val ARG_MODE = "mode"
        const val MODE_JOINED = "joined"
        private const val PAGE_SIZE = 30
    }
}
