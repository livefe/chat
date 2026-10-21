package com.nearby.chat.feature.map

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.KeyboardArrowDown
import androidx.compose.material.icons.outlined.LocationOn
import androidx.compose.material.icons.outlined.Search
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.Poi
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

private val SearchBarBorder = AppColors.BorderStrong
private val CityChipBorder = AppColors.BrandDarkBorder
private val SearchPlaceholderColor = AppColors.TextMuted
private val SearchBarHeight = 40.dp
private val CityChipMaxWidth = 128.dp
private val PillShape = RoundedCornerShape(20.dp)
private val PoiListMaxHeight = 260.dp
private val SmallIconSize = AppIconSize.Inline

@Composable
fun LocationPickSearchPanel(
    state: LocationPickUiState,
    onOpenCityPicker: () -> Unit,
    onKeyword: (String) -> Unit,
    onSearch: () -> Unit,
    onClearKeyword: () -> Unit,
    onSelectPoi: (Poi) -> Unit,
    modifier: Modifier = Modifier,
) {
    val focusManager = LocalFocusManager.current
    val cityLabel = state.selectedRegion?.name ?: stringResource(R.string.location_select_city)
    val searchPlaceholder = if (state.selectedRegion != null) {
        stringResource(R.string.location_search_in_city, cityLabel)
    } else {
        stringResource(R.string.location_search_place)
    }
    Column(
        modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.M, vertical = AppSpacing.S),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(
                modifier = Modifier
                    .height(SearchBarHeight)
                    .widthIn(max = CityChipMaxWidth)
                    .clip(PillShape)
                    .background(AppColors.BrandContainer)
                    .border(1.dp, CityChipBorder, PillShape)
                    .clickable {
                        focusManager.clearFocus()
                        onOpenCityPicker()
                    }
                    .padding(horizontal = AppSpacing.M),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(AppSpacing.Xs),
            ) {
                Icon(
                    Icons.Outlined.LocationOn,
                    contentDescription = null,
                    tint = AppColors.BrandDark,
                    modifier = Modifier.size(17.dp),
                )
                Text(
                    cityLabel,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    style = MaterialTheme.typography.labelLarge,
                    color = AppColors.BrandDark,
                    modifier = Modifier.weight(1f, fill = false),
                )
                Icon(
                    Icons.Outlined.KeyboardArrowDown,
                    contentDescription = null,
                    tint = AppColors.BrandDark,
                    modifier = Modifier.size(SmallIconSize),
                )
            }
            Row(
                modifier = Modifier
                    .weight(1f)
                    .height(SearchBarHeight)
                    .clip(PillShape)
                    .background(AppColors.CardBg)
                    .border(1.dp, SearchBarBorder, PillShape)
                    .padding(horizontal = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(
                    Icons.Outlined.Search,
                    contentDescription = null,
                    tint = AppColors.TextMuted,
                    modifier = Modifier.size(SmallIconSize),
                )
                BasicTextField(
                    value = state.keyword,
                    onValueChange = onKeyword,
                    modifier = Modifier
                        .weight(1f)
                        .imeOnTap()
                        .padding(horizontal = AppSpacing.S),
                    singleLine = true,
                    textStyle = TextStyle(
                        fontSize = AppType.Secondary,
                        color = AppColors.TextPrimary,
                    ),
                    cursorBrush = SolidColor(AppColors.Brand),
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                    keyboardActions = KeyboardActions(
                        onSearch = {
                            focusManager.clearFocus()
                            onSearch()
                        },
                    ),
                    decorationBox = { inner ->
                        Box(contentAlignment = Alignment.CenterStart) {
                            if (state.keyword.isEmpty()) {
                                Text(
                                    searchPlaceholder,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = SearchPlaceholderColor,
                                )
                            }
                            inner()
                        }
                    },
                )
                if (state.keyword.isNotEmpty()) {
                    Icon(
                        Icons.Outlined.Close,
                        contentDescription = null,
                        tint = AppColors.TextMuted,
                        modifier = Modifier
                            .size(SmallIconSize)
                            .clickable(onClick = onClearKeyword),
                    )
                }
            }
        }
        if (state.pois.isNotEmpty()) {
            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 6.dp)
                    .heightIn(max = PoiListMaxHeight)
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg),
            ) {
                items(state.pois) { poi ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                focusManager.clearFocus()
                                onSelectPoi(poi)
                            }
                            .padding(horizontal = 14.dp, vertical = AppSpacing.M),
                    ) {
                        Text(poi.name, style = MaterialTheme.typography.titleSmall)
                        if (poi.address.isNotBlank()) {
                            Text(
                                poi.address,
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    }
                    HorizontalDivider(color = AppColors.OverlaySoft)
                }
            }
        }
    }
}
