package com.nearby.chat.feature.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.text.ClickableText
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppCellField
import com.nearby.chat.core.ui.AppPlainPrimaryButton
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.LocalTextInputPressTracker
import com.nearby.chat.core.ui.TextInputPressTracker
import com.nearby.chat.core.ui.dismissKeyboardOnTap

private const val TAG_USER = "user"
private const val TAG_PRIVACY = "privacy"
private val LoginTopGap = 72.dp
private val LoginFormGap = 32.dp
private val AgreeRowOffset = (-30).dp
private val AgreeTouchSize = 44.dp
private val AgreeBoxSize = 14.dp
private val AgreeCheckSize = 10.dp

@Composable
fun LoginScreen(
    onLoggedIn: () -> Unit,
    onOpenUserAgreement: () -> Unit,
    onOpenPrivacyPolicy: () -> Unit,
    viewModel: LoginViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val focusManager = LocalFocusManager.current
    val keyboard = LocalSoftwareKeyboardController.current
    val textInputPress = remember { TextInputPressTracker() }
    val latestOnLoggedIn by rememberUpdatedState(onLoggedIn)

    fun dismissKeyboard() {
        focusManager.clearFocus(force = true)
        keyboard?.hide()
    }

    LaunchedEffect(viewModel) {
        viewModel.loggedIn.collect {
            dismissKeyboard()
            latestOnLoggedIn()
        }
    }

    val prefix = stringResource(R.string.login_agreement_prefix)
    val userAgreement = stringResource(R.string.login_agreement_user)
    val and = stringResource(R.string.login_agreement_and)
    val privacy = stringResource(R.string.login_agreement_privacy)
    val agreementText = remember(prefix, userAgreement, and, privacy) {
        buildAnnotatedString {
            append(prefix)
            pushStringAnnotation(tag = TAG_USER, annotation = TAG_USER)
            withStyle(SpanStyle(color = AppColors.Brand)) {
                append(userAgreement)
            }
            pop()
            append(and)
            pushStringAnnotation(tag = TAG_PRIVACY, annotation = TAG_PRIVACY)
            withStyle(SpanStyle(color = AppColors.Brand)) {
                append(privacy)
            }
            pop()
        }
    }

    CompositionLocalProvider(LocalTextInputPressTracker provides textInputPress) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(AppColors.PageBg)
                .statusBarsPadding()
                .dismissKeyboardOnTap()
                .padding(horizontal = AppSpacing.L),
        ) {
            Spacer(Modifier.height(LoginTopGap))
            Text(
                stringResource(R.string.app_name),
                fontSize = AppType.Hero,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.Brand,
            )
            Spacer(Modifier.height(AppSpacing.S))
            Text(
                stringResource(R.string.login_tagline),
                fontSize = AppType.Secondary,
                lineHeight = 20.sp,
                color = AppColors.TextTertiary,
            )

            Spacer(Modifier.height(LoginFormGap))

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(AppShapes.Form)
                    .background(AppColors.CardBg),
            ) {
                AppCellField(
                    label = stringResource(R.string.auth_mobile_label),
                    value = state.mobile,
                    onValueChange = viewModel::onMobileChange,
                    placeholder = stringResource(R.string.auth_mobile_hint),
                    keyboardType = KeyboardType.Phone,
                )
                HorizontalDivider(
                    color = AppColors.BorderSoft,
                    modifier = Modifier.padding(start = AppSpacing.L),
                )
                AppCellField(
                    label = stringResource(R.string.auth_code_label),
                    value = state.code,
                    onValueChange = viewModel::onCodeChange,
                    placeholder = stringResource(R.string.auth_code_hint),
                    keyboardType = KeyboardType.Number,
                    trailing = {
                        AppPlainPrimaryButton(
                            text = when {
                                state.countdown > 0 -> stringResource(R.string.auth_code_countdown, state.countdown)
                                state.sendingCode -> stringResource(R.string.auth_code_sending)
                                else -> stringResource(R.string.auth_get_code)
                            },
                            onClick = {
                                dismissKeyboard()
                                viewModel.sendCode()
                            },
                            enabled = state.countdown == 0 && !state.sendingCode,
                        )
                    },
                )
            }

            Spacer(Modifier.height(AppSpacing.Xxl))
            AppPrimaryButton(
                text = stringResource(if (state.loading) R.string.login_in_progress else R.string.login_action),
                onClick = {
                    dismissKeyboard()
                    viewModel.login()
                },
                enabled = !state.loading,
            )

            Spacer(Modifier.height(AppSpacing.L))
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier
                    .fillMaxWidth()
                    .offset(x = AgreeRowOffset),
            ) {
                LoginAgreeMark(
                    checked = state.agreed,
                    onToggle = viewModel::toggleAgreed,
                )
                ClickableText(
                    text = agreementText,
                    style = TextStyle(
                        fontSize = AppType.Meta,
                        color = AppColors.TextTertiary,
                        lineHeight = AppType.LineMeta,
                    ),
                    onClick = { offset ->
                        val ann = agreementText.getStringAnnotations(start = offset, end = offset)
                            .firstOrNull()
                        when (ann?.tag) {
                            TAG_USER -> onOpenUserAgreement()
                            TAG_PRIVACY -> onOpenPrivacyPolicy()
                            else -> viewModel.toggleAgreed()
                        }
                    },
                    modifier = Modifier
                        .weight(1f)
                        .padding(start = AppSpacing.S),
                )
            }
        }
    }
}

@Composable
private fun LoginAgreeMark(
    checked: Boolean,
    onToggle: () -> Unit,
) {
    val interaction = remember { MutableInteractionSource() }
    Box(
        contentAlignment = Alignment.CenterEnd,
        modifier = Modifier
            .size(AgreeTouchSize)
            .clickable(
                interactionSource = interaction,
                indication = null,
                role = Role.Checkbox,
                onClick = onToggle,
            ),
    ) {
        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier
                .size(AgreeBoxSize)
                .clip(AppShapes.Tag)
                .background(if (checked) AppColors.Brand else AppColors.CardBg)
                .border(
                    width = 1.dp,
                    color = if (checked) AppColors.Brand else AppColors.TextTertiary,
                    shape = AppShapes.Tag,
                ),
        ) {
            if (checked) {
                Icon(
                    imageVector = Icons.Filled.Check,
                    contentDescription = null,
                    tint = Color.White,
                    modifier = Modifier.size(AgreeCheckSize),
                )
            }
        }
    }
}
