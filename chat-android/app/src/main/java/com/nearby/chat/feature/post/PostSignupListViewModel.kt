package com.nearby.chat.feature.post

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.model.PostSignupItem
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.export.EmptySignupExportException
import com.nearby.chat.data.export.SignupExportRepository
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import javax.inject.Inject

@HiltViewModel
class PostSignupListViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val signupExportRepository: SignupExportRepository,
    private val toaster: Toaster,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val postId: Long = savedStateHandle.get<Long>(ARG_POST_ID) ?: 0L

    private val _uiState = MutableStateFlow(PostSignupListUiState(postId = postId))
    val uiState: StateFlow<PostSignupListUiState> = _uiState.asStateFlow()

    private val pager = OffsetPageTracker(PAGE_SIZE)

    init {
        if (postId > 0) {
            loadTitle()
            reload(showPullIndicator = false)
        }
    }

    fun setStatusFilter(status: Int?) {
        val current = _uiState.value
        if (current.statusFilter == status && current.checkInFilter == null) return
        _uiState.update { it.copy(statusFilter = status, checkInFilter = null) }
        reload(showPullIndicator = false)
    }

    fun setCheckInFilter(checkIn: Int) {
        val current = _uiState.value
        val next = if (current.checkInFilter == checkIn) null else checkIn
        if (current.statusFilter == 1 && current.checkInFilter == next) return
        _uiState.update { it.copy(statusFilter = 1, checkInFilter = next) }
        reload(showPullIndicator = false)
    }

    fun refresh() {
        reload(showPullIndicator = true)
    }

    fun loadMore() {
        val s = _uiState.value
        if (s.loading || !s.hasMore || s.refreshing || postId <= 0) return
        viewModelScope.launch { loadMoreInternal() }
    }

    fun export() {
        if (postId <= 0 || _uiState.value.exporting) return
        viewModelScope.launch {
            _uiState.update { it.copy(exporting = true) }
            toaster.show(R.string.signup_export_running)
            val s = _uiState.value
            runCatching {
                signupExportRepository.export(
                    postId = postId,
                    postTitle = s.postTitle,
                    status = s.statusFilter,
                    checkIn = s.checkInFilter,
                ) { needCheckIn, counts ->
                    _uiState.update { it.copy(needCheckIn = needCheckIn, counts = counts) }
                }
            }.onSuccess {
                toaster.success(R.string.signup_export_done)
            }.onFailure { e ->
                if (e is EmptySignupExportException) {
                    toaster.fail(R.string.signup_export_empty)
                } else {
                    toaster.fail(e, R.string.signup_export_failed)
                }
            }
            _uiState.update { it.copy(exporting = false) }
        }
    }

    fun manualCheckIn(userId: Long) {
        if (postId <= 0 || userId <= 0 || _uiState.value.checkingUserId != 0L) return
        viewModelScope.launch {
            _uiState.update { it.copy(checkingUserId = userId) }
            runCatching { postRepository.signupCheckIn(postId = postId, userId = userId) }
                .onSuccess { result ->
                    toaster.success(if (result.alreadyCheckedIn) R.string.signup_checked_in else R.string.check_in_success)
                    val checkedInAt = nowDateTime()
                    _uiState.update { s ->
                        val leaveUnchecked = s.checkInFilter == PostSignupRules.CHECK_IN_NOT_CHECKED
                        val counts = if (result.alreadyCheckedIn) {
                            s.counts
                        } else {
                            s.counts.copy(
                                notCheckedIn = (s.counts.notCheckedIn - 1).coerceAtLeast(0),
                                checkedIn = s.counts.checkedIn + 1,
                            )
                        }
                        s.copy(
                            checkingUserId = 0,
                            counts = counts,
                            list = if (leaveUnchecked && !result.alreadyCheckedIn) {
                                s.list.filter { it.userId != userId }
                            } else {
                                s.list.map { item ->
                                    if (item.userId == userId) item.copy(checkedInAt = checkedInAt) else item
                                }
                            },
                        )
                    }
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.check_in_failed)
                    _uiState.update { it.copy(checkingUserId = 0) }
                }
        }
    }

    fun openRemarkEditor(item: PostSignupItem) {
        if (_uiState.value.remarkSaving) return
        _uiState.update {
            it.copy(remarkEditing = item, remarkDraft = item.remark.trim())
        }
    }

    fun closeRemarkEditor() {
        if (_uiState.value.remarkSaving) return
        _uiState.update { it.copy(remarkEditing = null, remarkDraft = "") }
    }

    fun onRemarkDraftChange(value: String) {
        _uiState.update { it.copy(remarkDraft = value.take(REMARK_MAX)) }
    }

    fun saveRemark() {
        val target = _uiState.value.remarkEditing ?: return
        if (postId <= 0 || _uiState.value.remarkSaving) return
        val remark = _uiState.value.remarkDraft.trim()
        viewModelScope.launch {
            _uiState.update { it.copy(remarkSaving = true) }
            runCatching {
                postRepository.updateSignupRemark(postId, target.userId, remark)
            }.onSuccess {
                toaster.success(R.string.common_saved)
                _uiState.update { s ->
                    s.copy(
                        remarkSaving = false,
                        remarkEditing = null,
                        remarkDraft = "",
                        list = s.list.map { item ->
                            if (item.id == target.id) item.copy(remark = remark) else item
                        },
                    )
                }
            }.onFailure { e ->
                toaster.fail(e, R.string.common_save_failed)
                _uiState.update { it.copy(remarkSaving = false) }
            }
        }
    }

    private fun reload(showPullIndicator: Boolean) {
        if (postId <= 0) return
        viewModelScope.launch {
            pager.reset()
            _uiState.update {
                it.copy(
                    hasMore = true,
                    refreshing = showPullIndicator,
                    loading = !showPullIndicator,
                    list = if (showPullIndicator) it.list else emptyList(),
                )
            }
            loadMoreInternal(skipLoadingFlag = showPullIndicator)
            _uiState.update { it.copy(refreshing = false, loading = false) }
        }
    }

    private fun loadTitle() {
        viewModelScope.launch {
            runCatching { postRepository.detail(postId) }
                .onSuccess { detail ->
                    _uiState.update {
                        it.copy(
                            postTitle = detail.title,
                            needCheckIn = detail.extras?.needSignup == true &&
                                detail.extras.needCheckIn,
                            checkInWindowOpen = PostSignupRules.isCheckInWindowOpen(
                                detail.startAt,
                                detail.endAt,
                            ),
                        )
                    }
                }
        }
    }

    private suspend fun loadMoreInternal(skipLoadingFlag: Boolean = false) {
        if (!skipLoadingFlag) {
            _uiState.update { it.copy(loading = true) }
        }
        runCatching {
            postRepository.signupList(
                postId = postId,
                page = pager.page,
                pageSize = pager.pageSize,
                status = _uiState.value.statusFilter,
                checkIn = _uiState.value.checkInFilter,
            )
        }.onSuccess { data ->
            _uiState.update { state ->
                val result = pager.apply(state.list, data.total, data.list)
                state.copy(
                    list = result.list,
                    hasMore = result.hasMore,
                    loading = false,
                    needCheckIn = data.needCheckIn,
                    counts = data.counts,
                )
            }
        }.onFailure { e ->
            toaster.fail(e, R.string.common_load_failed)
            _uiState.update { it.copy(hasMore = false, loading = false) }
        }
    }

    companion object {
        const val ARG_POST_ID = "postId"
        const val REMARK_MAX = 100
        private const val PAGE_SIZE = 30

        private fun nowDateTime(): String {
            val format = SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault())
            return format.format(Date())
        }
    }
}
