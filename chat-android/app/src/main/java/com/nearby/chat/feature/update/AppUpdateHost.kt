package com.nearby.chat.feature.update

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle

@Composable
fun AppUpdateHost(viewModel: AppUpdateViewModel) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current

    LaunchedEffect(viewModel) {
        viewModel.checkOnLaunch()
    }
    DisposableEffect(lifecycleOwner, viewModel) {
        var firstResume = true
        val observer = LifecycleEventObserver { _, event ->
            if (event != Lifecycle.Event.ON_RESUME) return@LifecycleEventObserver
            if (firstResume) {
                firstResume = false
                return@LifecycleEventObserver
            }
            viewModel.checkOnResume()
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    state.prompt?.let { prompt ->
        AppUpdateDialog(
            data = prompt,
            onDismissSoft = viewModel::dismissSoft,
            onDownloadUnavailable = viewModel::onDownloadUnavailable,
        )
    }
}
