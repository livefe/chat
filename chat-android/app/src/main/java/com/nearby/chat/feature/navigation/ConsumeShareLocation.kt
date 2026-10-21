package com.nearby.chat.feature.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.navigation.NavBackStackEntry
import com.nearby.chat.feature.map.LocationPickResult

@Composable
internal fun ConsumeShareLocation(
    entry: NavBackStackEntry,
    onShare: (LocationPickResult) -> Unit,
) {
    val lifecycleOwner = LocalLifecycleOwner.current
    val latestOnShare by rememberUpdatedState(onShare)
    DisposableEffect(lifecycleOwner, entry) {
        val observer = LifecycleEventObserver { _, event ->
            if (event != Lifecycle.Event.ON_RESUME) return@LifecycleEventObserver
            val raw = entry.savedStateHandle.remove<String>(NavResultKeys.SHARE_LOCATION)
            val parsed = LocationPickResult.parseSharePayload(raw) ?: return@LifecycleEventObserver
            latestOnShare(parsed)
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }
}
