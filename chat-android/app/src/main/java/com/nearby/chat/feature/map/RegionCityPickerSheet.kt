package com.nearby.chat.feature.map

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.ArrowBack
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.Search
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.asString
import com.nearby.chat.core.ui.imeOnTap

private val RegionSearchHeight = 42.dp
private val RegionSearchShape = RoundedCornerShape(21.dp)

@Composable
fun RegionCityPickerSheet(
    visible: Boolean,
    onDismiss: () -> Unit,
    onSelect: (RegionItem) -> Unit,
    viewModel: RegionCityPickerViewModel = hiltViewModel(),
) {
    if (!visible) return
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val currentOnSelect by rememberUpdatedState(onSelect)
    val currentOnDismiss by rememberUpdatedState(onDismiss)

    LaunchedEffect(Unit) {
        viewModel.open()
    }
    LaunchedEffect(viewModel) {
        viewModel.selected.collect { region ->
            currentOnSelect(region)
            currentOnDismiss()
        }
    }

    RegionCityPickerSheetContent(
        state = state,
        onDismiss = onDismiss,
        onBack = viewModel::back,
        onKeyword = viewModel::onKeyword,
        onSearch = viewModel::search,
        onClearKeyword = viewModel::clearKeyword,
        onPick = viewModel::pick,
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RegionCityPickerSheetContent(
    state: RegionCityPickerUiState,
    onDismiss: () -> Unit,
    onBack: () -> Unit,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onClearKeyword: () -> Unit,
    onPick: (RegionItem) -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = AppColors.PageBg,
        modifier = Modifier.fillMaxHeight(0.72f),
    ) {
        RegionPickerHeader(canGoBack = state.stack.isNotEmpty(), onBack = onBack, onDismiss = onDismiss)
        RegionSearchField(
            keyword = state.keyword,
            onKeyword = onKeyword,
            onSearch = onSearch,
            onClear = onClearKeyword,
        )
        if (state.stack.isNotEmpty()) {
            RegionBreadcrumb(state.stack)
        } else {
            Spacer(modifier = Modifier.height(10.dp))
        }
        val error = state.error
        when {
            state.loading -> {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .heightIn(min = 160.dp)
                        .padding(AppSpacing.Xxl),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = AppColors.Brand)
                }
            }

            error != null -> {
                Text(
                    error.asString(),
                    color = AppColors.Danger,
                    modifier = Modifier.padding(AppSpacing.Xxl),
                )
            }

            state.list.isEmpty() -> {
                EmptyHint(stringResource(R.string.region_empty), Modifier.fillMaxWidth().heightIn(min = 220.dp))
            }

            else -> RegionList(list = state.list, onPick = onPick)
        }
    }
}

@Composable
private fun RegionPickerHeader(
    canGoBack: Boolean,
    onBack: () -> Unit,
    onDismiss: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp)
            .padding(horizontal = AppSpacing.Xs),
    ) {
        if (canGoBack) {
            TextButton(
                onClick = onBack,
                modifier = Modifier.align(Alignment.CenterStart),
                contentPadding = PaddingValues(horizontal = AppSpacing.S),
            ) {
                Icon(
                    Icons.AutoMirrored.Outlined.ArrowBack,
                    contentDescription = null,
                    tint = AppColors.BrandDark,
                    modifier = Modifier.size(AppIconSize.Inline),
                )
                Text(stringResource(R.string.common_back), color = AppColors.BrandDark)
            }
            Text(
                stringResource(R.string.location_select_city),
                modifier = Modifier.align(Alignment.Center),
                fontSize = AppSheetTokens.TitleSize,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.TextPrimary,
            )
        } else {
            Text(
                stringResource(R.string.location_select_city),
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = AppSpacing.M),
                fontSize = AppSheetTokens.TitleSize,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.TextPrimary,
            )
        }
        IconButton(
            onClick = onDismiss,
            modifier = Modifier.align(Alignment.CenterEnd),
        ) {
            Icon(
                Icons.Outlined.Close,
                contentDescription = stringResource(R.string.common_close),
                tint = AppColors.TextSecondary,
            )
        }
    }
}

@Composable
private fun RegionSearchField(
    keyword: String,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onClear: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L)
            .height(RegionSearchHeight)
            .clip(RegionSearchShape)
            .background(AppColors.CardBg)
            .border(1.dp, AppColors.NeutralTint, RegionSearchShape)
            .padding(horizontal = AppSpacing.M),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            Icons.Outlined.Search,
            contentDescription = null,
            tint = AppColors.TextSecondary,
            modifier = Modifier.size(AppIconSize.Inline),
        )
        BasicTextField(
            value = keyword,
            onValueChange = onKeyword,
            modifier = Modifier
                .weight(1f)
                .imeOnTap()
                .padding(horizontal = AppSpacing.S),
            singleLine = true,
            textStyle = TextStyle(fontSize = AppType.Secondary, color = AppColors.TextPrimary),
            cursorBrush = SolidColor(AppColors.Brand),
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
            keyboardActions = KeyboardActions(onSearch = { onSearch() }),
            decorationBox = { inner ->
                Box(contentAlignment = Alignment.CenterStart) {
                    if (keyword.isEmpty()) {
                        Text(
                            stringResource(R.string.region_search_hint),
                            color = AppColors.TextSecondary,
                            fontSize = AppType.Caption,
                        )
                    }
                    inner()
                }
            },
        )
        if (keyword.isNotEmpty()) {
            Icon(
                Icons.Outlined.Close,
                contentDescription = null,
                tint = AppColors.TextSecondary,
                modifier = Modifier
                    .size(AppIconSize.Inline)
                    .clickable(onClick = onClear),
            )
        }
    }
}

@Composable
private fun RegionBreadcrumb(stack: List<RegionItem>) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L, vertical = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        stack.forEachIndexed { index, item ->
            if (index > 0) {
                Text("/", color = AppColors.TextSecondary, fontSize = AppType.Caption)
            }
            Text(
                item.name,
                color = if (index == stack.lastIndex) AppColors.BrandDark else AppColors.TextSecondary,
                fontSize = AppType.Caption,
                fontWeight = if (index == stack.lastIndex) FontWeight.Medium else FontWeight.Normal,
            )
        }
    }
}

@Composable
private fun RegionList(
    list: List<RegionItem>,
    onPick: (RegionItem) -> Unit,
) {
    LazyColumn(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.M)
            .clip(AppShapes.Card)
            .background(AppColors.CardBg),
    ) {
        items(list, key = { it.code }) { item ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onPick(item) }
                    .padding(horizontal = AppSpacing.L, vertical = AppSheetTokens.ItemVPad),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    item.name,
                    modifier = Modifier.weight(1f),
                    fontSize = AppSheetTokens.ItemSize,
                    fontWeight = FontWeight.Medium,
                    color = AppColors.TextPrimary,
                )
                if (item.level == 1) {
                    Icon(
                        Icons.AutoMirrored.Outlined.KeyboardArrowRight,
                        contentDescription = null,
                        tint = AppColors.TextTertiary,
                    )
                }
            }
            HorizontalDivider(color = AppColors.OverlaySoft)
        }
    }
}
