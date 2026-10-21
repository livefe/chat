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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CheckboxDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppCellField
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppPlainPrimaryButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

private val DangerButtonHeight = 44.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AccountDeactivateScreen(
    onBack: () -> Unit,
    onDone: () -> Unit,
    viewModel: AccountDeactivateViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnDone by rememberUpdatedState(onDone)

    LaunchedEffect(viewModel) {
        viewModel.deactivated.collect { latestOnDone() }
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.deactivate_title)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(bottom = 32.dp),
        ) {
            Column(
                modifier = Modifier
                    .padding(AppSpacing.L)
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg)
                    .padding(horizontal = AppSpacing.L, vertical = 14.dp),
            ) {
                Text(
                    stringResource(R.string.deactivate_notice_title),
                    fontSize = AppType.Title,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                )
                Spacer(Modifier.height(AppSpacing.S))
                Text(
                    stringResource(R.string.deactivate_notice_body),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                    lineHeight = 22.sp,
                )
            }

            Column(
                modifier = Modifier
                    .padding(horizontal = AppSpacing.L)
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg),
            ) {
                AppCellField(
                    label = stringResource(R.string.auth_mobile_label),
                    value = state.mobile,
                    readOnly = true,
                    placeholder = stringResource(R.string.common_placeholder_dash),
                )
                HorizontalDivider(
                    color = AppColors.Divider,
                    modifier = Modifier.padding(start = AppSpacing.L),
                )
                AppCellField(
                    label = stringResource(R.string.auth_code_label),
                    value = state.code,
                    onValueChange = viewModel::onCode,
                    placeholder = stringResource(R.string.auth_code_hint),
                    keyboardType = KeyboardType.Number,
                    trailing = {
                        AppPlainPrimaryButton(
                            text = when {
                                state.sending -> stringResource(R.string.auth_code_sending)
                                state.counting > 0 -> stringResource(R.string.auth_code_countdown, state.counting)
                                else -> stringResource(R.string.auth_get_code)
                            },
                            onClick = viewModel::sendCode,
                            enabled = state.counting == 0 && !state.sending,
                        )
                    },
                )
            }

            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(horizontal = AppSpacing.M, vertical = AppSpacing.L),
            ) {
                Checkbox(
                    checked = state.agreed,
                    onCheckedChange = viewModel::onAgreed,
                    colors = CheckboxDefaults.colors(checkedColor = AppColors.Brand),
                )
                Text(
                    stringResource(R.string.deactivate_agree),
                    fontSize = AppType.Caption,
                    color = AppColors.TextSecondary,
                    modifier = Modifier.clickable { viewModel.onAgreed(!state.agreed) },
                )
            }

            Button(
                onClick = viewModel::requestConfirm,
                enabled = !state.submitting,
                shape = AppShapes.Pill,
                colors = ButtonDefaults.buttonColors(
                    containerColor = AppColors.Danger,
                    disabledContainerColor = AppColors.Danger.copy(alpha = 0.4f),
                ),
                modifier = Modifier
                    .padding(horizontal = AppSpacing.L)
                    .fillMaxWidth()
                    .height(DangerButtonHeight),
            ) {
                Text(
                    stringResource(if (state.submitting) R.string.deactivate_submitting else R.string.deactivate_submit),
                    fontSize = AppType.Subhead,
                )
            }
        }
    }

    if (state.confirmVisible) {
        AppConfirmDialog(
            title = stringResource(R.string.deactivate_confirm_title),
            message = stringResource(R.string.deactivate_confirm_message),
            confirmText = stringResource(R.string.common_ok),
            cancelText = stringResource(R.string.common_cancel),
            danger = true,
            onConfirm = viewModel::submit,
            onDismiss = viewModel::dismissConfirm,
        )
    }
}
