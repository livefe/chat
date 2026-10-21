package com.nearby.chat.feature.update

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.AppUpdateRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class AppUpdateViewModel @Inject constructor(
    private val repo: AppUpdateRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(AppUpdateUiState())
    val uiState: StateFlow<AppUpdateUiState> = _uiState.asStateFlow()

    fun checkOnLaunch() {
        check(manual = false, respectSoftThrottle = false)
    }

    fun checkOnResume() {
        check(manual = false, respectSoftThrottle = true)
    }

    fun checkManual() {
        check(manual = true, respectSoftThrottle = false)
    }

    fun dismissSoft() {
        val prompt = _uiState.value.prompt ?: return
        if (!prompt.forceUpdate) {
            repo.skipSoftUpdate(prompt.latestVersion)
            repo.markSoftPrompted()
        }
        _uiState.update { it.copy(prompt = null) }
    }

    fun onDownloadUnavailable() {
        toaster.fail(R.string.update_no_download_url)
    }

    private fun check(manual: Boolean, respectSoftThrottle: Boolean) {
        viewModelScope.launch {
            _uiState.update { it.copy(checking = true) }
            runCatching { repo.check() }
                .onSuccess { data ->
                    if (!data.hasUpdate && !data.forceUpdate) {
                        _uiState.update { it.copy(checking = false, prompt = null) }
                        if (manual) {
                            if (data.latestVersion.isBlank()) {
                                toaster.show(R.string.update_latest)
                            } else {
                                toaster.show(R.string.update_latest_version, data.latestVersion)
                            }
                        }
                        return@onSuccess
                    }
                    if (data.forceUpdate) {
                        _uiState.update { it.copy(checking = false, prompt = data) }
                        return@onSuccess
                    }

                    if (!manual) {
                        if (repo.skippedSoftVersion() == data.latestVersion) {
                            _uiState.update { it.copy(checking = false, prompt = null) }
                            return@onSuccess
                        }
                        if (respectSoftThrottle) {
                            val elapsed = System.currentTimeMillis() - repo.lastSoftPromptAt()
                            if (elapsed < AppUpdateRepository.SOFT_PROMPT_INTERVAL_MS) {
                                _uiState.update { it.copy(checking = false, prompt = null) }
                                return@onSuccess
                            }
                        }
                        repo.markSoftPrompted()
                    }
                    _uiState.update { it.copy(checking = false, prompt = data) }
                }
                .onFailure {
                    _uiState.update {
                        it.copy(
                            checking = false,
                            prompt = if (manual) it.prompt else null,
                        )
                    }
                    if (manual) toaster.show(R.string.update_check_failed)
                }
        }
    }
}
