package com.nearby.chat.feature.feedback

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
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
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.model.FeedbackItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

private val ProgressStroke = 2.dp
private val ImageThumbSize = 80.dp
private val DetailRowPadding = 10.dp
private val DetailBlockLineHeight = 22.sp
private const val LABEL_WEIGHT = 0.35f
private const val VALUE_WEIGHT = 0.65f

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedbackDetailScreen(
    onBack: () -> Unit,
    viewModel: FeedbackDetailViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(state.loading, state.detail) {
        if (!state.loading && state.detail == null) onBack()
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.feedback_detail_title)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        val detail = state.detail
        when {
            state.loading -> {
                Box(
                    modifier = Modifier
                        .padding(padding)
                        .fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = AppColors.Brand, strokeWidth = ProgressStroke)
                }
            }

            detail != null -> {
                FeedbackDetailContent(
                    detail = detail,
                    modifier = Modifier
                        .padding(padding)
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState())
                        .padding(bottom = AppSpacing.Xxl),
                )
            }
        }
    }
}

@Composable
private fun FeedbackDetailContent(detail: FeedbackItem, modifier: Modifier) {
    Column(modifier = modifier) {
        DetailCard {
            DetailRow(stringResource(R.string.feedback_detail_type), detail.typeText)
            HorizontalDivider(color = AppColors.Divider)
            DetailRow(
                stringResource(R.string.feedback_detail_status),
                detail.statusText,
                valueColor = feedbackStatusColor(detail.status),
            )
            HorizontalDivider(color = AppColors.Divider)
            DetailRow(stringResource(R.string.feedback_detail_created_at), detail.createdAt)
            HorizontalDivider(color = AppColors.Divider)
            DetailBlock(stringResource(R.string.feedback_content_label), detail.content)
            if (detail.contact.isNotBlank()) {
                HorizontalDivider(color = AppColors.Divider)
                DetailRow(stringResource(R.string.feedback_contact_label), detail.contact)
            }
        }

        if (detail.images.isNotEmpty()) {
            Spacer(Modifier.height(AppSpacing.M))
            DetailCard {
                Text(
                    stringResource(R.string.feedback_detail_images),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                    modifier = Modifier.padding(bottom = DetailRowPadding),
                )
                Row(horizontalArrangement = Arrangement.spacedBy(AppSpacing.S)) {
                    detail.images.forEach { url ->
                        AsyncImage(
                            model = MediaUrl.resolve(url),
                            contentDescription = null,
                            contentScale = ContentScale.Crop,
                            modifier = Modifier
                                .size(ImageThumbSize)
                                .clip(AppShapes.Media)
                                .background(AppColors.PageBg),
                        )
                    }
                }
            }
        }

        Spacer(Modifier.height(AppSpacing.M))
        DetailCard {
            Text(
                stringResource(R.string.feedback_detail_result),
                fontSize = AppType.Secondary,
                color = AppColors.TextSecondary,
                modifier = Modifier.padding(bottom = AppSpacing.S),
            )
            if (detail.handleResult.isNotBlank() || detail.handledAt.isNotBlank()) {
                if (detail.handleResult.isNotBlank()) {
                    DetailBlock(stringResource(R.string.feedback_detail_reply), detail.handleResult)
                }
                if (detail.handledAt.isNotBlank()) {
                    if (detail.handleResult.isNotBlank()) {
                        HorizontalDivider(
                            color = AppColors.Divider,
                            modifier = Modifier.padding(vertical = AppSpacing.S),
                        )
                    }
                    DetailRow(stringResource(R.string.feedback_detail_handled_at), detail.handledAt)
                }
            } else {
                Text(
                    stringResource(R.string.feedback_detail_pending),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextPrimary,
                )
            }
        }
    }
}

@Composable
private fun DetailCard(content: @Composable () -> Unit) {
    Column(
        modifier = Modifier
            .padding(horizontal = AppSpacing.L)
            .fillMaxWidth()
            .clip(AppShapes.Card)
            .background(AppColors.CardBg)
            .padding(AppSpacing.L),
        content = { content() },
    )
}

@Composable
private fun DetailRow(
    label: String,
    value: String,
    valueColor: Color = AppColors.TextPrimary,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = DetailRowPadding),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(label, fontSize = AppType.Secondary, color = AppColors.TextSecondary, modifier = Modifier.weight(LABEL_WEIGHT))
        Text(value, fontSize = AppType.Secondary, color = valueColor, modifier = Modifier.weight(VALUE_WEIGHT))
    }
}

@Composable
private fun DetailBlock(label: String, value: String) {
    Column(modifier = Modifier.padding(vertical = DetailRowPadding)) {
        Text(label, fontSize = AppType.Secondary, color = AppColors.TextSecondary)
        Text(
            value,
            fontSize = AppType.Secondary,
            color = AppColors.TextPrimary,
            lineHeight = DetailBlockLineHeight,
            modifier = Modifier.padding(top = 6.dp),
        )
    }
}
