package com.nearby.chat.feature.account

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.format.maskMobile
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.imeOnTap

private val UnderlineFieldHeight = 48.dp
private val ActionTopGap = 40.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditMobileScreen(
    onBack: () -> Unit,
    viewModel: EditMobileViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnBack by rememberUpdatedState(onBack)
    LaunchedEffect(Unit) { viewModel.load() }
    LaunchedEffect(viewModel) {
        viewModel.saved.collect { latestOnBack() }
    }

    fun handleBack() {
        if (!viewModel.onBack()) onBack()
    }

    BackHandler { handleBack() }

    val verifyStep = state.step == EditMobileUiState.STEP_VERIFY_OLD
    val title = stringResource(if (verifyStep) R.string.mobile_verify_title else R.string.mobile_bind_title)
    val masked = maskMobile(state.currentMobile).ifBlank { stringResource(R.string.common_placeholder_dash) }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(title) },
                navigationIcon = {
                    AppBackIconButton(onClick = { handleBack() })
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
                .padding(horizontal = AppSpacing.Xl)
                .padding(top = 28.dp, bottom = 32.dp),
        ) {
            if (verifyStep) {
                Text(
                    stringResource(R.string.mobile_verify_desc),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextTertiary,
                    lineHeight = 21.sp,
                )
                Text(
                    masked,
                    fontSize = AppType.Hero,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                    letterSpacing = 1.sp,
                    modifier = Modifier.padding(top = AppSpacing.Xl, bottom = 28.dp),
                )
                UnderlineCodeRow(
                    value = state.oldCode,
                    onValueChange = viewModel::onOldCode,
                    placeholder = stringResource(R.string.auth_code_hint),
                    codeLabel = codeLabel(state.sendingOld, state.oldCounting),
                    codeEnabled = state.oldCounting == 0 && !state.sendingOld,
                    onSendCode = viewModel::sendOldCode,
                )
                AppPrimaryButton(
                    text = stringResource(if (state.verifying) R.string.mobile_verifying else R.string.mobile_next),
                    onClick = viewModel::goNext,
                    enabled = state.oldCode.length >= EditMobileViewModel.CODE_MIN_LENGTH &&
                        !state.sendingOld &&
                        !state.verifying,
                    modifier = Modifier.padding(top = ActionTopGap),
                )
            } else {
                Text(
                    stringResource(R.string.mobile_bind_desc),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextTertiary,
                    lineHeight = 21.sp,
                    modifier = Modifier.padding(bottom = AppSpacing.Xl),
                )
                UnderlineField(
                    value = state.newMobile,
                    onValueChange = viewModel::onNewMobile,
                    placeholder = stringResource(R.string.mobile_new_hint),
                    keyboardType = KeyboardType.Phone,
                )
                UnderlineCodeRow(
                    value = state.newCode,
                    onValueChange = viewModel::onNewCode,
                    placeholder = stringResource(R.string.auth_code_hint),
                    codeLabel = codeLabel(state.sendingNew, state.newCounting),
                    codeEnabled = state.newCounting == 0 && !state.sendingNew,
                    onSendCode = viewModel::sendNewCode,
                    modifier = Modifier.padding(top = AppSpacing.S),
                )
                AppPrimaryButton(
                    text = stringResource(if (state.submitting) R.string.mobile_changing else R.string.mobile_change_confirm),
                    onClick = viewModel::submit,
                    enabled = state.ticket.isNotBlank() &&
                        state.newMobile.length == EditMobileViewModel.MOBILE_LENGTH &&
                        state.newCode.length >= EditMobileViewModel.CODE_MIN_LENGTH &&
                        !state.submitting,
                    modifier = Modifier.padding(top = ActionTopGap),
                )
            }
        }
    }
}

@Composable
private fun codeLabel(sending: Boolean, counting: Int): String = when {
    sending -> stringResource(R.string.mobile_code_sending)
    counting > 0 -> stringResource(R.string.auth_code_countdown, counting)
    else -> stringResource(R.string.auth_get_code)
}

@Composable
private fun UnderlineField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    keyboardType: KeyboardType,
    modifier: Modifier = Modifier,
) {
    Column(modifier = modifier.fillMaxWidth()) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(UnderlineFieldHeight),
            contentAlignment = Alignment.CenterStart,
        ) {
            BasicTextField(
                value = value,
                onValueChange = onValueChange,
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
                textStyle = TextStyle(fontSize = AppType.PageTitle, color = AppColors.TextPrimary),
                cursorBrush = SolidColor(AppColors.Brand),
                modifier = Modifier.fillMaxWidth().imeOnTap(),
                decorationBox = { inner ->
                    if (value.isEmpty()) {
                        Text(placeholder, color = AppColors.TextTertiary, fontSize = AppType.PageTitle)
                    }
                    inner()
                },
            )
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(AppColors.BorderSoft),
        )
    }
}

@Composable
private fun UnderlineCodeRow(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    codeLabel: String,
    codeEnabled: Boolean,
    onSendCode: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(modifier = modifier.fillMaxWidth()) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier
                .fillMaxWidth()
                .height(UnderlineFieldHeight),
        ) {
            BasicTextField(
                value = value,
                onValueChange = onValueChange,
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                textStyle = TextStyle(fontSize = AppType.PageTitle, color = AppColors.TextPrimary),
                cursorBrush = SolidColor(AppColors.Brand),
                modifier = Modifier.weight(1f).imeOnTap(),
                decorationBox = { inner ->
                    if (value.isEmpty()) {
                        Text(placeholder, color = AppColors.TextTertiary, fontSize = AppType.PageTitle)
                    }
                    inner()
                },
            )
            TextButton(
                onClick = onSendCode,
                enabled = codeEnabled,
            ) {
                Text(
                    codeLabel,
                    color = if (codeEnabled) AppColors.Brand else AppColors.TextTertiary,
                    fontSize = AppType.Subhead,
                )
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(AppColors.BorderSoft),
        )
    }
}
