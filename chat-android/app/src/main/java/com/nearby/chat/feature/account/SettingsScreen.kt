package com.nearby.chat.feature.account

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import com.nearby.chat.BuildConfig
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    onBack: () -> Unit,
    onOpenBlocks: () -> Unit,
    onOpenDeactivate: () -> Unit,
    onCheckUpdate: () -> Unit,
    checkingUpdate: Boolean = false,
) {
    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.me_settings)) },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg),
            ) {
                SettingsCell(stringResource(R.string.block_list_title), onClick = onOpenBlocks)
                HorizontalDivider(color = AppColors.Divider, modifier = Modifier.padding(start = AppSpacing.L))
                SettingsCell(stringResource(R.string.deactivate_title), onClick = onOpenDeactivate)
            }

            Spacer(Modifier.height(AppSpacing.M))

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg),
            ) {
                SettingsCell(
                    title = stringResource(R.string.settings_check_update),
                    value = if (checkingUpdate) {
                        stringResource(R.string.settings_checking_update)
                    } else {
                        BuildConfig.VERSION_NAME
                    },
                    onClick = onCheckUpdate,
                )
            }
        }
    }
}

@Composable
private fun SettingsCell(
    title: String,
    onClick: () -> Unit,
    value: String? = null,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(AppSpacing.L),
    ) {
        Text(
            title,
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            modifier = Modifier.weight(1f),
        )
        if (value != null) {
            Text(
                value,
                fontSize = AppType.Body,
                color = AppColors.TextSecondary,
                modifier = Modifier.padding(end = AppSpacing.Xs),
            )
        }
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextSecondary,
            modifier = Modifier.size(AppIconSize.Cell),
        )
    }
}
