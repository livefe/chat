package com.nearby.chat.feature.discover

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.model.AdEvent
import com.nearby.chat.core.model.AdEventType
import com.nearby.chat.core.model.FeedItem
import com.nearby.chat.core.model.SponsoredItem
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AdRepository
import com.nearby.chat.data.repo.LocationRepository
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class DiscoverViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val adRepository: AdRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(DiscoverUiState())
    val uiState: StateFlow<DiscoverUiState> = _uiState.asStateFlow()

    private val pager = OffsetPageTracker(PAGE_SIZE)
    private var loadJob: Job? = null
    private val impressedTokens = mutableSetOf<String>()

    fun refresh() {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            pager.reset()
            impressedTokens.clear()
            _uiState.update {
                it.copy(
                    hasMore = true,
                    refreshing = true,
                    loading = false,
                )
            }
            loadMoreInternal(showLoading = false)
            _uiState.update { it.copy(refreshing = false, loading = false) }
        }
    }

    fun load() {
        if (_uiState.value.list.isNotEmpty()) {
            refresh()
            return
        }
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            pager.reset()
            impressedTokens.clear()
            _uiState.update { it.copy(hasMore = true, loading = true) }
            loadMoreInternal()
        }
    }

    fun loadMore() {
        val s = _uiState.value
        if (s.loading || !s.hasMore || s.refreshing) return
        loadJob = viewModelScope.launch { loadMoreInternal() }
    }

    fun reportSponsoredImpression(item: SponsoredItem) {
        val token = item.impressionToken
        if (token.isBlank() || !impressedTokens.add(token)) return
        viewModelScope.launch {
            runCatching { reportAdEvent(item, AdEventType.IMPRESSION) }
                .onFailure { impressedTokens.remove(token) }
        }
    }

    fun onSponsoredClick(item: SponsoredItem) {
        viewModelScope.launch {
            runCatching { reportAdEvent(item, AdEventType.CLICK) }
        }
    }

    private suspend fun reportAdEvent(item: SponsoredItem, eventType: String) {
        val saved = runCatching { locationRepository.savedLocation() }.getOrNull()
        adRepository.reportEvents(
            listOf(
                AdEvent(
                    requestId = item.requestId,
                    campaignId = item.campaignId,
                    eventType = eventType,
                    impressionToken = item.impressionToken,
                    slotIndex = item.slotIndex,
                    latitude = saved?.lat ?: 0.0,
                    longitude = saved?.lng ?: 0.0,
                ),
            ),
        )
    }

    private suspend fun loadMoreInternal(showLoading: Boolean = true) {
        if (showLoading) {
            _uiState.update { it.copy(loading = true) }
        }
        val saved = runCatching { locationRepository.savedLocation() }.getOrNull()
        runCatching {
            postRepository.list(
                page = pager.page,
                pageSize = pager.pageSize,
                latitude = saved?.lat,
                longitude = saved?.lng,
            )
        }.onSuccess { (total, items) ->
            _uiState.update { state ->
                val result = pager.apply(
                    existing = state.list,
                    total = total,
                    items = items,
                    countedSize = { list -> list.count { it is FeedItem.Post } },
                )
                state.copy(list = result.list, hasMore = result.hasMore, loading = false)
            }
        }.onFailure { e ->
            if (e is CancellationException) throw e
            toaster.fail(e, R.string.common_load_failed)
            _uiState.update { it.copy(hasMore = false, loading = false) }
        }
    }

    companion object {
        private const val PAGE_SIZE = 20
    }
}
