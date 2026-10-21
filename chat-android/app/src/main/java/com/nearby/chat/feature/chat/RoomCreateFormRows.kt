package com.nearby.chat.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Slider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

private val SliderTrackInactive = AppColors.NeutralLine
private val SliderThumbShadow = AppColors.ShadowSoft
private val SliderThumbBorder = AppColors.NeutralTint
internal val RoomFormLabelWidth = 72.dp
internal val RoomFormRowVPad = 14.dp
private val FieldLineHeight = 20.sp
private val SliderHeight = 28.dp
private val SliderThumbSize = 20.dp
private val SliderTrackHeight = 2.dp
private const val RADIUS_STEP = 50f

@Composable
internal fun RoomFormDivider() {
    HorizontalDivider(color = AppColors.Border, modifier = Modifier.padding(start = AppSpacing.L))
}

@Composable
internal fun RoomFormField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    singleLine: Boolean,
    minLines: Int = 1,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L, vertical = RoomFormRowVPad),
        verticalAlignment = if (singleLine) Alignment.CenterVertically else Alignment.Top,
    ) {
        Text(
            label,
            fontSize = AppType.Secondary,
            color = AppColors.TextPrimary,
            modifier = Modifier
                .width(RoomFormLabelWidth)
                .padding(top = if (singleLine) 0.dp else AppSpacing.Xxs),
        )
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = singleLine,
            minLines = minLines,
            textStyle = TextStyle(
                color = AppColors.TextPrimary,
                fontSize = AppType.Secondary,
                lineHeight = FieldLineHeight,
            ),
            cursorBrush = SolidColor(AppColors.Brand),
            modifier = Modifier.weight(1f).imeOnTap(),
            decorationBox = { inner ->
                if (value.isEmpty()) {
                    Text(placeholder, color = AppColors.TextTertiary, fontSize = AppType.Secondary)
                }
                inner()
            },
        )
    }
}

@Composable
internal fun RoomFormLinkRow(
    label: String,
    value: String,
    placeholder: Boolean,
    onClick: () -> Unit,
    maxLines: Int = 1,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = AppSpacing.L, vertical = RoomFormRowVPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            label,
            fontSize = AppType.Secondary,
            color = AppColors.TextPrimary,
            modifier = Modifier.width(RoomFormLabelWidth),
        )
        Text(
            value,
            fontSize = AppType.Secondary,
            color = if (placeholder) AppColors.TextTertiary else AppColors.TextPrimary,
            maxLines = maxLines,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.End,
            modifier = Modifier.weight(1f),
        )
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
            modifier = Modifier.size(AppIconSize.Inline),
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun RoomRadiusSlider(
    value: Float,
    onValueChange: (Float) -> Unit,
    valueRange: ClosedFloatingPointRange<Float>,
    modifier: Modifier = Modifier,
) {
    val interactionSource = remember { MutableInteractionSource() }
    val steps = ((valueRange.endInclusive - valueRange.start) / RADIUS_STEP).toInt() - 1
    Slider(
        value = value,
        onValueChange = { raw ->
            val stepped = ((raw - valueRange.start) / RADIUS_STEP).toInt() * RADIUS_STEP + valueRange.start
            onValueChange(stepped.coerceIn(valueRange.start, valueRange.endInclusive))
        },
        valueRange = valueRange,
        steps = steps,
        interactionSource = interactionSource,
        modifier = modifier.height(SliderHeight),
        thumb = {
            Box(
                modifier = Modifier
                    .size(SliderThumbSize)
                    .shadow(3.dp, CircleShape, ambientColor = SliderThumbShadow, spotColor = SliderThumbShadow)
                    .border(0.5.dp, SliderThumbBorder, CircleShape)
                    .background(AppColors.Surface, CircleShape),
            )
        },
        track = { sliderState ->
            val range = sliderState.valueRange
            val fraction = (
                (sliderState.value - range.start) /
                    (range.endInclusive - range.start).coerceAtLeast(0.0001f)
                ).coerceIn(0f, 1f)
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(SliderHeight),
                contentAlignment = Alignment.CenterStart,
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(SliderTrackHeight)
                        .clip(CircleShape)
                        .background(SliderTrackInactive),
                )
                Box(
                    modifier = Modifier
                        .fillMaxWidth(fraction)
                        .height(SliderTrackHeight)
                        .clip(CircleShape)
                        .background(AppColors.Brand),
                )
            }
        },
    )
}
