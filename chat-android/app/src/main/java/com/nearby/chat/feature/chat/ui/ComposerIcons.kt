package com.nearby.chat.feature.chat.ui

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.path
import androidx.compose.ui.unit.dp

private const val STROKE_WIDTH = 1.5f

val ComposerMicOutlined: ImageVector by lazy(LazyThreadSafetyMode.NONE) { buildComposerMicOutlined() }

private fun buildComposerMicOutlined(): ImageVector = ImageVector.Builder(
    name = "ComposerMicOutlined",
    defaultWidth = 24.dp,
    defaultHeight = 24.dp,
    viewportWidth = 24f,
    viewportHeight = 24f,
).apply {
    path(
        fill = null,
        stroke = SolidColor(Color.Black),
        strokeLineWidth = STROKE_WIDTH,
        strokeLineCap = StrokeCap.Round,
        strokeLineJoin = StrokeJoin.Round,
    ) {
        moveTo(12f, 2.5f)
        curveTo(9.7909f, 2.5f, 8f, 4.2909f, 8f, 6.5f)
        verticalLineTo(10.5f)
        curveTo(8f, 12.7091f, 9.7909f, 14.5f, 12f, 14.5f)
        curveTo(14.2091f, 14.5f, 16f, 12.7091f, 16f, 10.5f)
        verticalLineTo(6.5f)
        curveTo(16f, 4.2909f, 14.2091f, 2.5f, 12f, 2.5f)
        close()
    }
    path(
        fill = null,
        stroke = SolidColor(Color.Black),
        strokeLineWidth = STROKE_WIDTH,
        strokeLineCap = StrokeCap.Round,
        strokeLineJoin = StrokeJoin.Round,
    ) {
        moveTo(6.5f, 12.5f)
        curveTo(6.5f, 15.5376f, 8.9624f, 18f, 12f, 18f)
        curveTo(15.0376f, 18f, 17.5f, 15.5376f, 17.5f, 12.5f)
        moveTo(12f, 18f)
        verticalLineTo(21.25f)
        moveTo(9.5f, 21.25f)
        horizontalLineTo(14.5f)
    }
}.build()
