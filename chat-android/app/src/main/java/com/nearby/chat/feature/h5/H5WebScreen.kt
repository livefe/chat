package com.nearby.chat.feature.h5

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.asString

private val ErrorTopPadding = 48.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun H5WebScreen(
    url: String,
    title: String = "",
    onBack: () -> Unit,
) {
    val context = LocalContext.current
    val defaultTitle = stringResource(R.string.h5_default_title)
    val controller = remember(title) { H5WebController(context, title, defaultTitle) }
    val latestOnBack by rememberUpdatedState(onBack)

    DisposableEffect(controller) {
        controller.attach(onClose = { latestOnBack() })
        onDispose { controller.release() }
    }

    LaunchedEffect(controller, url) { controller.load(url) }

    val goBack = { if (!controller.goBack()) onBack() }

    BackHandler(onBack = goBack)

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(controller.pageTitle) },
                navigationIcon = { AppBackIconButton(onClick = goBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            AndroidView(
                factory = { controller.webView },
                modifier = Modifier.fillMaxSize(),
            )
            val loadError = controller.loadError
            if (controller.progress in 0f..<1f && loadError == null) {
                LinearProgressIndicator(
                    progress = { controller.progress },
                    modifier = Modifier
                        .fillMaxWidth()
                        .align(Alignment.TopCenter),
                    color = AppColors.Brand,
                    trackColor = AppColors.BorderSoft,
                )
            }
            if (loadError != null) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(AppSpacing.Xxl),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Text(
                        loadError.asString(),
                        color = AppColors.TextSecondary,
                        fontSize = AppType.Secondary,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(top = ErrorTopPadding),
                    )
                    TextButton(onClick = controller::retry) {
                        Text(stringResource(R.string.common_retry))
                    }
                }
            }
        }
    }
}
