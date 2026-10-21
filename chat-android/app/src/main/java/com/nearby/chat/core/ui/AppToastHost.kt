package com.nearby.chat.core.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.BiasAlignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import kotlinx.coroutines.delay

private val ToastBg = AppColors.ToastBg
private val ToastSuccess = AppColors.Success
private val ToastFail = AppColors.Danger

private val ToastAlignment = BiasAlignment(horizontalBias = 0f, verticalBias = -0.35f)

private data class ShownToast(
    val message: String,
    val kind: ToastKind,
)

@Composable
fun AppToastHost(toaster: Toaster = LocalToaster.current) {
    val context = LocalContext.current
    var current by remember { mutableStateOf<ShownToast?>(null) }
    var visible by remember { mutableStateOf(false) }

    LaunchedEffect(toaster) {
        toaster.events.collect { toast ->
            val text = toast.message.resolve(context).trim()
            if (text.isEmpty()) return@collect
            val event = ShownToast(text, toast.kind)
            current = event
            visible = true
            delay(2200)
            visible = false
            delay(180)
            if (current == event) current = null
        }
    }

    Box(modifier = Modifier.fillMaxSize(), contentAlignment = ToastAlignment) {
        AnimatedVisibility(
            visible = visible && current != null,
            enter = fadeIn() + scaleIn(initialScale = 0.92f),
            exit = fadeOut() + scaleOut(targetScale = 0.92f),
        ) {
            val event = current ?: return@AnimatedVisibility
            val accent = when (event.kind) {
                ToastKind.Success -> ToastSuccess
                ToastKind.Fail -> ToastFail
                ToastKind.Info -> Color.Transparent
            }
            Box(
                modifier = Modifier
                    .padding(horizontal = 40.dp)
                    .widthIn(min = 96.dp, max = 280.dp)
                    .clip(AppShapes.Pill)
                    .background(ToastBg)
                    .then(
                        if (accent != Color.Transparent) {
                            Modifier.padding(top = AppSpacing.Xxs)
                        } else {
                            Modifier
                        },
                    )
                    .padding(horizontal = AppSpacing.M, vertical = AppSpacing.S),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = event.message,
                    color = Color.White,
                    fontSize = AppType.Secondary,
                    textAlign = TextAlign.Center,
                    lineHeight = 20.sp,
                )
            }
        }
    }
}
