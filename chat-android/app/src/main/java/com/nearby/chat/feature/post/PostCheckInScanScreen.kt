package com.nearby.chat.feature.post

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.content.pm.ActivityInfo
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.isGranted
import com.google.accompanist.permissions.rememberPermissionState
import com.google.accompanist.permissions.shouldShowRationale
import com.nearby.chat.R
import com.nearby.chat.core.device.Haptics
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPrimaryButton

private const val HAPTIC_SUCCESS_MS = 60L
private const val HAPTIC_ALREADY_MS = 40L
private val HapticFailPattern = longArrayOf(0, 40, 60, 40)
private val TopScrimHeight = 96.dp

@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun PostCheckInScanScreen(
    onBack: () -> Unit,
    onCheckedIn: (CheckInNavResult) -> Unit,
    viewModel: PostCheckInScanViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val haptics = remember(context) { Haptics(context) }
    val cameraPermission = rememberPermissionState(Manifest.permission.CAMERA)
    val latestOnCheckedIn by rememberUpdatedState(onCheckedIn)

    DisposableEffect(Unit) {
        val activity = context.findActivity()
        val previous = activity?.requestedOrientation
        activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        onDispose {
            if (previous != null) activity.requestedOrientation = previous
        }
    }

    LaunchedEffect(Unit) {
        if (!cameraPermission.status.isGranted) {
            cameraPermission.launchPermissionRequest()
        }
    }

    LaunchedEffect(viewModel) {
        viewModel.results.collect { result ->
            when (result.kind) {
                CheckInNavResult.KIND_FAIL -> haptics.pattern(HapticFailPattern)
                CheckInNavResult.KIND_ALREADY -> haptics.tick(HAPTIC_ALREADY_MS)
                else -> haptics.tick(HAPTIC_SUCCESS_MS)
            }
            latestOnCheckedIn(result)
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (cameraPermission.status.isGranted) {
            EmbeddedQrScanner(
                paused = state.busy,
                onScanned = viewModel::checkIn,
                modifier = Modifier.fillMaxSize(),
            )
            ScanSceneOverlay(
                tip = stringResource(state.tip),
                modifier = Modifier.fillMaxSize(),
            )
            Box(
                modifier = Modifier
                    .align(Alignment.TopCenter)
                    .fillMaxWidth()
                    .height(TopScrimHeight)
                    .background(
                        Brush.verticalGradient(
                            listOf(
                                Color.Black.copy(alpha = 0.32f),
                                Color.Transparent,
                            ),
                        ),
                    ),
            )
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .statusBarsPadding()
                    .padding(horizontal = AppSpacing.Xxl),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                Text(
                    stringResource(
                        if (cameraPermission.status.shouldShowRationale) {
                            R.string.scan_need_camera
                        } else {
                            R.string.scan_need_camera_settings
                        },
                    ),
                    color = Color.White.copy(alpha = 0.75f),
                    fontSize = AppType.Secondary,
                    textAlign = TextAlign.Center,
                )
                AppPrimaryButton(
                    text = stringResource(R.string.scan_grant),
                    onClick = { cameraPermission.launchPermissionRequest() },
                    fillMaxWidth = false,
                    compact = true,
                    modifier = Modifier.padding(top = AppSpacing.L),
                )
            }
        }
        ScanTopBar(onBack = onBack)
    }
}

@Composable
private fun BoxScope.ScanTopBar(onBack: () -> Unit) {
    Row(
        modifier = Modifier
            .align(Alignment.TopStart)
            .fillMaxWidth()
            .statusBarsPadding()
            .padding(AppSpacing.Xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        AppBackIconButton(onClick = onBack)
        Text(
            stringResource(R.string.scan_action),
            fontSize = AppType.PageTitle,
            fontWeight = FontWeight.Medium,
            color = Color.White.copy(alpha = 0.96f),
            modifier = Modifier.padding(start = AppSpacing.Xs),
        )
    }
}

private fun Context.findActivity(): Activity? {
    var current: Context? = this
    while (current is ContextWrapper) {
        if (current is Activity) return current
        current = current.baseContext
    }
    return null
}
