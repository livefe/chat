package com.nearby.chat.feature.post

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathFillType
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.clipRect
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private const val FRAME_RATIO = 0.68f
private const val CORNER_RATIO = 0.14f
private const val TRAIL_RATIO = 0.36f
private const val BEAM_DURATION_MS = 2400

@Composable
internal fun ScanSceneOverlay(
    tip: String,
    modifier: Modifier = Modifier,
) {
    val transition = rememberInfiniteTransition(label = "scanBeam")
    val beamProgress by transition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(BEAM_DURATION_MS, easing = LinearEasing),
            repeatMode = RepeatMode.Restart,
        ),
        label = "beam",
    )
    val corner = AppColors.BrandLight.copy(alpha = 0.85f)
    val mask = Color.Black.copy(alpha = 0.42f)

    Box(modifier = modifier) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            val frame = size.minDimension * FRAME_RATIO
            val left = (size.width - frame) / 2f
            val top = (size.height - frame) / 2f
            val right = left + frame
            val bottom = top + frame

            val hole = Path().apply {
                fillType = PathFillType.EvenOdd
                addRect(Rect(0f, 0f, size.width, size.height))
                addRect(Rect(left, top, right, bottom))
            }
            drawPath(hole, mask)

            val stroke = 2.5.dp.toPx()
            val len = frame * CORNER_RATIO
            val cap = StrokeCap.Round
            drawLine(corner, Offset(left, top), Offset(left + len, top), stroke, cap)
            drawLine(corner, Offset(left, top), Offset(left, top + len), stroke, cap)
            drawLine(corner, Offset(right - len, top), Offset(right, top), stroke, cap)
            drawLine(corner, Offset(right, top), Offset(right, top + len), stroke, cap)
            drawLine(corner, Offset(left, bottom), Offset(left + len, bottom), stroke, cap)
            drawLine(corner, Offset(left, bottom - len), Offset(left, bottom), stroke, cap)
            drawLine(corner, Offset(right - len, bottom), Offset(right, bottom), stroke, cap)
            drawLine(corner, Offset(right, bottom - len), Offset(right, bottom), stroke, cap)

            clipRect(left, top, right, bottom) {
                val trailH = frame * TRAIL_RATIO
                val lineH = 2.dp.toPx()
                val y = top + (frame + trailH) * beamProgress - trailH
                drawRect(
                    brush = Brush.verticalGradient(
                        colorStops = arrayOf(
                            0f to Color.Transparent,
                            0.4f to AppColors.Brand.copy(alpha = 0.06f),
                            0.75f to AppColors.Brand.copy(alpha = 0.18f),
                            0.93f to AppColors.BrandLight.copy(alpha = 0.38f),
                            1f to AppColors.Brand.copy(alpha = 0.1f),
                        ),
                        startY = y,
                        endY = y + trailH,
                    ),
                    topLeft = Offset(left, y),
                    size = Size(frame, trailH),
                )
                drawRect(
                    brush = Brush.horizontalGradient(
                        colors = listOf(
                            Color.Transparent,
                            AppColors.BrandLight.copy(alpha = 0.6f),
                            Color.White.copy(alpha = 0.95f),
                            AppColors.BrandLight.copy(alpha = 0.6f),
                            Color.Transparent,
                        ),
                        startX = left,
                        endX = right,
                    ),
                    topLeft = Offset(left, y + trailH - lineH),
                    size = Size(frame, lineH),
                )
            }
        }

        Text(
            tip,
            color = Color.White.copy(alpha = 0.92f),
            fontSize = AppType.Subhead,
            textAlign = TextAlign.Center,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .background(
                    Brush.verticalGradient(
                        listOf(
                            Color.Transparent,
                            Color.Black.copy(alpha = 0.28f),
                            Color.Black.copy(alpha = 0.5f),
                        ),
                    ),
                )
                .padding(horizontal = AppSpacing.Xl, vertical = AppSpacing.L),
        )
    }
}
