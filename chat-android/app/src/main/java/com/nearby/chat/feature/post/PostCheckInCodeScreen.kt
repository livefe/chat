package com.nearby.chat.feature.post

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

private val QrBoxSize = 240.dp
private const val CHECKED_IN_QR_ALPHA = 0.55f

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostCheckInCodeScreen(
    onBack: () -> Unit,
    viewModel: PostCheckInCodeViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val codeLabel = stringResource(R.string.check_in_code_label)

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(codeLabel) },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = AppSpacing.Xl, vertical = AppSpacing.Xxl),
            contentAlignment = Alignment.TopCenter,
        ) {
            if (state.loading) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                }
                return@Box
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .verticalScroll(rememberScrollState())
                    .background(AppColors.CardBg, AppShapes.CardLg)
                    .padding(horizontal = AppSpacing.Xl, vertical = 28.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    state.postTitle.ifBlank { stringResource(R.string.check_in_code_default_title) },
                    fontSize = AppType.PageTitle,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                    textAlign = TextAlign.Center,
                )
                Text(
                    stringResource(if (state.checkedIn) R.string.check_in_code_hint_done else R.string.check_in_code_hint),
                    fontSize = AppType.Meta,
                    color = AppColors.TextTertiary,
                    modifier = Modifier.padding(top = AppSpacing.S),
                )
                Box(
                    modifier = Modifier
                        .padding(top = AppSpacing.Xxl)
                        .size(QrBoxSize)
                        .background(Color.White, AppShapes.Card)
                        .padding(AppSpacing.M)
                        .alpha(if (state.checkedIn) CHECKED_IN_QR_ALPHA else 1f),
                    contentAlignment = Alignment.Center,
                ) {
                    val qr = state.qrBitmap
                    if (qr != null) {
                        Image(
                            bitmap = qr.asImageBitmap(),
                            contentDescription = codeLabel,
                            modifier = Modifier.fillMaxSize(),
                        )
                    }
                }
                Text(
                    stringResource(
                        if (state.checkedIn) R.string.signup_checked_in else R.string.signup_badge_check_in_pending,
                    ),
                    fontSize = AppType.Subhead,
                    fontWeight = FontWeight.SemiBold,
                    color = if (state.checkedIn) AppColors.SuccessStrong else AppColors.TextSecondary,
                    modifier = Modifier.padding(top = AppSpacing.Xl),
                )
                if (state.checkedIn && state.checkedInAtText.isNotBlank()) {
                    Text(
                        stringResource(R.string.check_in_time, state.checkedInAtText),
                        fontSize = AppType.Meta,
                        color = AppColors.TextTertiary,
                        modifier = Modifier.padding(top = 6.dp),
                    )
                }
                if (state.userNickname.isNotBlank() || state.signupAt.isNotBlank()) {
                    HorizontalDivider(
                        color = AppColors.Divider,
                        modifier = Modifier.padding(top = AppSpacing.Xl),
                    )
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = AppSpacing.L),
                        verticalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        if (state.userNickname.isNotBlank()) {
                            InfoRow(label = stringResource(R.string.signup_nickname), value = state.userNickname)
                        }
                        if (state.signupAt.isNotBlank()) {
                            InfoRow(label = stringResource(R.string.signup_export_col_signup_at), value = state.signupAt)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, fontSize = AppType.Meta, color = AppColors.TextTertiary)
        Text(
            value,
            fontSize = AppType.Meta,
            color = AppColors.TextPrimary,
            textAlign = TextAlign.End,
            modifier = Modifier
                .weight(1f, fill = false)
                .padding(start = AppSpacing.M),
        )
    }
}
