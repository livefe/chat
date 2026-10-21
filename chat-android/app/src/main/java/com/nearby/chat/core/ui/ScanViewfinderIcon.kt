package com.nearby.chat.core.ui

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.path
import androidx.compose.ui.unit.dp

private var cachedScanViewfinderIcon: ImageVector? = null

val ScanViewfinderIcon: ImageVector
    get() {
        cachedScanViewfinderIcon?.let { return it }
        return ImageVector.Builder(
            name = "ScanViewfinder",
            defaultWidth = 24.dp,
            defaultHeight = 24.dp,
            viewportWidth = 24f,
            viewportHeight = 24f,
        ).apply {
            path(
                fill = SolidColor(Color.Transparent),
                stroke = SolidColor(Color.Black),
                strokeLineWidth = 1.75f,
                strokeLineCap = StrokeCap.Round,
                strokeLineJoin = StrokeJoin.Round,
            ) {
                moveTo(5f, 9f)
                lineTo(5f, 5f)
                lineTo(9f, 5f)
                moveTo(15f, 5f)
                lineTo(19f, 5f)
                lineTo(19f, 9f)
                moveTo(19f, 15f)
                lineTo(19f, 19f)
                lineTo(15f, 19f)
                moveTo(9f, 19f)
                lineTo(5f, 19f)
                lineTo(5f, 15f)
            }
            path(
                fill = SolidColor(Color.Transparent),
                stroke = SolidColor(Color.Black),
                strokeLineWidth = 1.5f,
                strokeLineCap = StrokeCap.Round,
            ) {
                moveTo(7f, 12f)
                lineTo(17f, 12f)
            }
        }.build().also { cachedScanViewfinderIcon = it }
    }
