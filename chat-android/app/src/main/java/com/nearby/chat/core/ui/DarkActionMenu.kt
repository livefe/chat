package com.nearby.chat.core.ui

import android.view.View
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.IntRect
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.LayoutDirection
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Popup
import androidx.compose.ui.window.PopupPositionProvider
import androidx.compose.ui.window.PopupProperties
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppType
import kotlin.math.roundToInt

private val MenuBg = AppColors.MenuDarkBg
private val MenuShadow = AppColors.MenuShadow
private val MenuShape = AppShapes.Pill

data class DarkActionItem(
    val label: String,
    val onClick: () -> Unit,
)

@Composable
private fun ClearPopupWindowBackground() {
    val view = LocalView.current
    SideEffect {
        var current: View? = view
        while (current != null) {
            current.setBackgroundColor(Color.Transparent.toArgb())
            val parent = current.parent as? View ?: break
            current = parent
            if (current.id == android.R.id.content) {
                current.setBackgroundColor(Color.Transparent.toArgb())
                break
            }
        }
    }
}

@Composable
fun DarkActionPill(actions: List<DarkActionItem>) {
    if (actions.isEmpty()) return
    ClearPopupWindowBackground()
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .shadow(
                elevation = 6.dp,
                shape = MenuShape,
                clip = false,
                ambientColor = MenuShadow,
                spotColor = MenuShadow,
            )
            .clip(MenuShape)
            .background(MenuBg),
    ) {
        actions.forEach { item ->
            Text(
                text = item.label,
                color = Color.White,
                fontSize = AppType.Meta,
                fontWeight = FontWeight.Medium,
                modifier = Modifier
                    .clickable(onClick = item.onClick)
                    .padding(horizontal = 14.dp, vertical = 10.dp),
            )
        }
    }
}

@Composable
fun DarkActionPillPopup(
    anchor: Rect,
    actions: List<DarkActionItem>,
    onDismiss: () -> Unit,
    focusable: Boolean = true,
    dismissOnClickOutside: Boolean = focusable,
) {
    if (actions.isEmpty()) return
    val anchorState = remember { mutableStateOf(anchor) }
    anchorState.value = anchor
    val positionProvider = remember {
        object : PopupPositionProvider {
            override fun calculatePosition(
                anchorBounds: IntRect,
                windowSize: IntSize,
                layoutDirection: LayoutDirection,
                popupContentSize: IntSize,
            ): IntOffset {
                val a = anchorState.value
                val margin = 8
                val gap = 8
                val maxX = (windowSize.width - popupContentSize.width - margin).coerceAtLeast(margin)
                val x = (a.center.x - popupContentSize.width / 2f).roundToInt()
                    .coerceIn(margin, maxX)
                val yAbove = a.top - popupContentSize.height - gap
                val yBelow = a.bottom + gap
                val maxY = (windowSize.height - popupContentSize.height - margin).coerceAtLeast(0)
                val placeBelow = yAbove < margin &&
                    yBelow + popupContentSize.height <= windowSize.height - margin
                val y = (if (placeBelow) yBelow else yAbove).roundToInt().coerceIn(0, maxY)
                return IntOffset(x, y)
            }
        }
    }
    Popup(
        popupPositionProvider = positionProvider,
        onDismissRequest = if (focusable || dismissOnClickOutside) onDismiss else ({}),
        properties = PopupProperties(
            focusable = focusable,
            dismissOnBackPress = focusable,
            dismissOnClickOutside = dismissOnClickOutside,
            clippingEnabled = false,
        ),
    ) {
        DarkActionPill(actions)
    }
}
