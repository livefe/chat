package com.nearby.chat.feature.map

import androidx.annotation.StringRes
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.repo.LocationRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RegionCityPickerViewModel @Inject constructor(
    private val locationRepository: LocationRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(RegionCityPickerUiState())
    val uiState: StateFlow<RegionCityPickerUiState> = _uiState.asStateFlow()

    private val _selected = MutableSharedFlow<RegionItem>(extraBufferCapacity = 1)
    val selected: SharedFlow<RegionItem> = _selected.asSharedFlow()

    private var job: Job? = null

    fun open() {
        _uiState.update { it.copy(keyword = "", stack = emptyList()) }
        launchLoad { loadChildren("") }
    }

    fun onKeyword(value: String) {
        _uiState.update { it.copy(keyword = value) }
        launchLoad {
            delay(300)
            runSearch(value)
        }
    }

    fun search() {
        val keyword = _uiState.value.keyword
        launchLoad { runSearch(keyword) }
    }

    fun clearKeyword() {
        _uiState.update { it.copy(keyword = "") }
        launchLoad { runSearch("") }
    }

    fun back() {
        _uiState.update { it.copy(stack = it.stack.dropLast(1), keyword = "") }
        val parent = _uiState.value.stack.lastOrNull()?.code.orEmpty()
        launchLoad { loadChildren(parent) }
    }

    fun pick(item: RegionItem) {
        if (item.level != 1) {
            _selected.tryEmit(item)
            return
        }
        launchLoad {
            _uiState.update { it.copy(loading = true) }
            runCatching { locationRepository.regionChildren(item.code) }
                .onSuccess { children ->
                    if (children.isEmpty()) {
                        _selected.tryEmit(item)
                    } else {
                        _uiState.update { it.copy(stack = it.stack + item, keyword = "", list = children) }
                    }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.errorText(R.string.region_load_failed)) }
                }
            _uiState.update { it.copy(loading = false) }
        }
    }

    private fun launchLoad(block: suspend () -> Unit) {
        job?.cancel()
        job = viewModelScope.launch { block() }
    }

    private suspend fun loadChildren(parentCode: String) {
        _uiState.update { it.copy(loading = true, error = null) }
        runCatching { locationRepository.regionChildren(parentCode) }
            .onSuccess { list -> _uiState.update { it.copy(list = list) } }
            .onFailure { e ->
                _uiState.update { it.copy(list = emptyList(), error = e.errorText(R.string.region_load_failed)) }
            }
        _uiState.update { it.copy(loading = false) }
    }

    private suspend fun runSearch(text: String) {
        val q = text.trim()
        if (q.isEmpty()) {
            loadChildren(_uiState.value.stack.lastOrNull()?.code.orEmpty())
            return
        }
        _uiState.update { it.copy(loading = true, error = null) }
        runCatching { locationRepository.searchRegions(keyword = q, limit = 30) }
            .onSuccess { list -> _uiState.update { it.copy(list = list) } }
            .onFailure { e ->
                _uiState.update { it.copy(list = emptyList(), error = e.errorText(R.string.common_search_failed)) }
            }
        _uiState.update { it.copy(loading = false) }
    }

    private fun Throwable.errorText(@StringRes fallback: Int): UiText = message?.let { UiText.of(it) } ?: UiText.of(fallback)
}
