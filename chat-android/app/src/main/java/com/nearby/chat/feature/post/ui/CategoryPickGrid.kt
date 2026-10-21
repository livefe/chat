package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.post.PostCategories
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private const val GRID_COLUMNS = 2

@Composable
fun CategoryPickGrid(
    selected: String,
    onSelect: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(AppSpacing.M),
    ) {
        PostCategories.options.chunked(GRID_COLUMNS).forEach { row ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(AppSpacing.M),
            ) {
                row.forEach { opt ->
                    val active = selected == opt.value
                    Column(
                        modifier = Modifier
                            .weight(1f)
                            .clip(AppShapes.Card)
                            .border(
                                width = 1.dp,
                                color = if (active) AppColors.Brand else AppColors.BorderSoft,
                                shape = AppShapes.Card,
                            )
                            .background(if (active) AppColors.BrandContainer else AppColors.CardBg)
                            .clickable { onSelect(opt.value) }
                            .padding(horizontal = AppSpacing.L, vertical = 18.dp),
                    ) {
                        Text(stringResource(opt.label), fontSize = AppType.Body, fontWeight = FontWeight.SemiBold)
                        Text(
                            stringResource(opt.hint),
                            fontSize = AppType.Caption,
                            color = AppColors.TextTertiary,
                            modifier = Modifier.padding(top = 6.dp),
                        )
                    }
                }
                if (row.size < GRID_COLUMNS) {
                    Spacer(Modifier.weight(1f))
                }
            }
        }
    }
}
