package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.snapping.rememberSnapFlingBehavior
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import kotlinx.coroutines.flow.distinctUntilChanged
import java.util.Calendar
import kotlin.math.abs

private val WHEEL_ITEM_HEIGHT = 40.dp
private const val WHEEL_VISIBLE_COUNT = 5

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DateTimeWheelSheet(
    visible: Boolean,
    title: String = stringResource(R.string.datetime_pick_title),
    initialSeconds: Long = 0,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
    onChange: (Long) -> Unit,
) {
    if (!visible) return

    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    val initial = remember(initialSeconds, visible) {
        Calendar.getInstance().apply {
            if (initialSeconds > 0) timeInMillis = initialSeconds * 1000
            set(Calendar.SECOND, 0)
            set(Calendar.MILLISECOND, 0)
        }
    }
    val onChangeUpdated by rememberUpdatedState(onChange)

    val nowYear = Calendar.getInstance().get(Calendar.YEAR)
    val years = remember(nowYear) { ((nowYear - 1)..(nowYear + 5)).toList() }
    val months = remember { (1..12).toList() }
    val hours = remember { (0..23).toList() }
    val minutes = remember { (0..59).toList() }

    var year by remember(initial) { mutableIntStateOf(initial.get(Calendar.YEAR)) }
    var month by remember(initial) { mutableIntStateOf(initial.get(Calendar.MONTH) + 1) }
    var day by remember(initial) { mutableIntStateOf(initial.get(Calendar.DAY_OF_MONTH)) }
    var hour by remember(initial) { mutableIntStateOf(initial.get(Calendar.HOUR_OF_DAY)) }
    var minute by remember(initial) { mutableIntStateOf(initial.get(Calendar.MINUTE)) }

    val daysInMonth = remember(year, month) { daysInMonth(year, month) }
    val days = remember(daysInMonth) { (1..daysInMonth).toList() }
    LaunchedEffect(daysInMonth) {
        if (day > daysInMonth) day = daysInMonth
    }

    LaunchedEffect(year, month, day, hour, minute) {
        val safeDay = day.coerceIn(1, daysInMonth(year, month))
        val cal = Calendar.getInstance().apply {
            set(Calendar.YEAR, year)
            set(Calendar.MONTH, month - 1)
            set(Calendar.DAY_OF_MONTH, safeDay)
            set(Calendar.HOUR_OF_DAY, hour)
            set(Calendar.MINUTE, minute)
            set(Calendar.SECOND, 0)
            set(Calendar.MILLISECOND, 0)
        }
        onChangeUpdated(cal.timeInMillis / 1000)
    }

    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = AppSpacing.S, vertical = AppSpacing.Xs),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TextButton(onClick = onDismiss) {
                Text(
                    stringResource(R.string.common_cancel),
                    color = AppColors.TextTertiary,
                    fontSize = AppSheetTokens.ActionSize,
                )
            }
            Text(
                title,
                modifier = Modifier.weight(1f),
                textAlign = TextAlign.Center,
                fontSize = AppSheetTokens.TitleSize,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.TextPrimary,
            )
            TextButton(onClick = onConfirm) {
                Text(
                    stringResource(R.string.common_confirm),
                    color = AppColors.BrandDark,
                    fontSize = AppSheetTokens.ActionSize,
                    fontWeight = FontWeight.SemiBold,
                )
            }
        }
        HorizontalDivider(color = AppColors.BorderSoft)

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(WHEEL_ITEM_HEIGHT * WHEEL_VISIBLE_COUNT)
                .padding(bottom = AppSpacing.M),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            WheelColumn(
                items = years.map { stringResource(R.string.datetime_year, it) },
                selectedIndex = years.indexOf(year).coerceAtLeast(0),
                onSelect = { year = years[it] },
                modifier = Modifier.weight(1.15f),
            )
            WheelColumn(
                items = months.map { stringResource(R.string.datetime_month, it) },
                selectedIndex = (month - 1).coerceIn(0, 11),
                onSelect = { month = months[it] },
                modifier = Modifier.weight(0.85f),
            )
            WheelColumn(
                items = days.map { stringResource(R.string.datetime_day, it) },
                selectedIndex = (day - 1).coerceIn(0, days.lastIndex),
                onSelect = { day = days[it] },
                modifier = Modifier.weight(0.85f),
            )
            Box(
                modifier = Modifier
                    .width(1.dp)
                    .fillMaxHeight()
                    .padding(vertical = AppSpacing.M)
                    .background(AppColors.BorderSoft),
            )
            WheelColumn(
                items = hours.map { stringResource(R.string.datetime_hour, it.toString().padStart(2, '0')) },
                selectedIndex = hour.coerceIn(0, 23),
                onSelect = { hour = hours[it] },
                modifier = Modifier.weight(0.9f),
            )
            WheelColumn(
                items = minutes.map { stringResource(R.string.datetime_minute, it.toString().padStart(2, '0')) },
                selectedIndex = minute.coerceIn(0, 59),
                onSelect = { minute = minutes[it] },
                modifier = Modifier.weight(0.9f),
            )
        }
        Spacer(modifier = Modifier.height(AppSpacing.S))
    }
}

@Composable
private fun WheelColumn(
    items: List<String>,
    selectedIndex: Int,
    onSelect: (Int) -> Unit,
    modifier: Modifier = Modifier,
) {
    val listState = rememberLazyListState()
    val fling = rememberSnapFlingBehavior(lazyListState = listState)
    val onSelectUpdated by rememberUpdatedState(onSelect)
    val sidePad = WHEEL_ITEM_HEIGHT * ((WHEEL_VISIBLE_COUNT - 1) / 2)

    LaunchedEffect(items) {
        val target = selectedIndex.coerceIn(0, (items.size - 1).coerceAtLeast(0))
        listState.scrollToItem(target)
    }

    LaunchedEffect(listState, items) {
        snapshotFlow {
            val layout = listState.layoutInfo
            if (layout.visibleItemsInfo.isEmpty()) return@snapshotFlow -1
            val viewportCenter = layout.viewportStartOffset + layout.viewportSize.height / 2f
            layout.visibleItemsInfo.minByOrNull { info ->
                abs((info.offset + info.size / 2f) - viewportCenter)
            }?.index ?: -1
        }
            .distinctUntilChanged()
            .collect { index ->
                if (index in items.indices) onSelectUpdated(index)
            }
    }

    Box(modifier = modifier.fillMaxHeight()) {
        Box(
            modifier = Modifier
                .align(Alignment.Center)
                .fillMaxWidth()
                .height(WHEEL_ITEM_HEIGHT)
                .background(AppColors.BrandContainer),
        )
        LazyColumn(
            state = listState,
            flingBehavior = fling,
            contentPadding = PaddingValues(vertical = sidePad),
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            items(items.size, key = { "${it}_${items[it]}" }) { index ->
                val selected = index == selectedIndex
                Text(
                    text = items[index],
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(WHEEL_ITEM_HEIGHT)
                        .padding(horizontal = AppSpacing.Xxs),
                    textAlign = TextAlign.Center,
                    fontSize = if (selected) AppType.Title else AppType.Secondary,
                    fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
                    color = if (selected) AppColors.TextPrimary else AppColors.TextTertiary,
                )
            }
        }
        WheelEdgeFade(Alignment.TopCenter)
        WheelEdgeFade(Alignment.BottomCenter)
    }
}

@Composable
private fun BoxScope.WheelEdgeFade(alignment: Alignment) {
    val brush = if (alignment == Alignment.TopCenter) {
        Brush.verticalGradient(listOf(AppColors.CardBg, Color.Transparent))
    } else {
        Brush.verticalGradient(listOf(Color.Transparent, AppColors.CardBg))
    }
    Box(
        modifier = Modifier
            .align(alignment)
            .fillMaxWidth()
            .height(WHEEL_ITEM_HEIGHT * 1.2f)
            .background(brush),
    )
}

private fun daysInMonth(year: Int, month: Int): Int {
    val cal = Calendar.getInstance()
    cal.set(Calendar.YEAR, year)
    cal.set(Calendar.MONTH, month - 1)
    cal.set(Calendar.DAY_OF_MONTH, 1)
    return cal.getActualMaximum(Calendar.DAY_OF_MONTH)
}
