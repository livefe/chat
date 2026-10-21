package com.nearby.chat.feature.report

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppFilterChip
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.ImageUploadGrid
import com.nearby.chat.core.ui.ImageUploadItem
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.imeOnTap

private val CardPadding = 14.dp
private val InnerGap = 10.dp
private val TargetAvatarSize = 40.dp
private val ContentHeight = 120.dp

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun ReportScreen(
    targetType: Int,
    targetId: Long,
    previewTitle: String = "",
    previewAvatar: String = "",
    onBack: () -> Unit,
    viewModel: ReportViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnBack by rememberUpdatedState(onBack)
    val targetLabel = stringResource(ReportTargetType.label(targetType))
    val displayTitle = previewTitle.trim().ifBlank { targetLabel }
    val pickImages = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickMultipleVisualMedia(ReportViewModel.MAX_IMAGES),
    ) { uris ->
        if (uris.isNotEmpty()) viewModel.addImages(uris)
    }

    LaunchedEffect(viewModel) {
        viewModel.submitted.collect { latestOnBack() }
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.report_title)) },
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
                .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
        ) {
            ReportCard {
                Text(
                    stringResource(R.string.report_target_prefix, targetLabel),
                    fontSize = AppType.Caption,
                    color = AppColors.TextTertiary,
                )
                Spacer(Modifier.height(InnerGap))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    if (targetType == ReportTargetType.USER) {
                        UserAvatar(
                            url = previewAvatar.takeIf { it.isNotBlank() },
                            size = TargetAvatarSize,
                        )
                        Spacer(Modifier.size(InnerGap))
                    }
                    Text(
                        text = displayTitle,
                        fontSize = AppType.Subhead,
                        fontWeight = FontWeight.Medium,
                        color = AppColors.TextPrimary,
                        maxLines = 2,
                    )
                }
            }

            Spacer(Modifier.height(AppSpacing.M))

            ReportCard {
                Text(
                    stringResource(R.string.report_reason_label),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextPrimary,
                    fontWeight = FontWeight.Medium,
                )
                Spacer(Modifier.height(InnerGap))
                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    verticalArrangement = Arrangement.spacedBy(AppSpacing.S),
                ) {
                    ReportReason.entries.forEach { reason ->
                        AppFilterChip(
                            label = stringResource(reason.label),
                            selected = state.reason == reason.value,
                            onClick = { viewModel.onReason(reason.value) },
                        )
                    }
                }
            }

            Spacer(Modifier.height(AppSpacing.M))

            ReportCard {
                Text(
                    stringResource(R.string.report_content_label),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                )
                OutlinedTextField(
                    value = state.content,
                    onValueChange = viewModel::onContent,
                    placeholder = {
                        Text(stringResource(R.string.report_content_hint), color = AppColors.TextTertiary)
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .imeOnTap()
                        .padding(top = AppSpacing.S)
                        .height(ContentHeight),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = AppColors.Brand,
                        unfocusedBorderColor = AppColors.BorderSoft,
                        focusedContainerColor = Color.Transparent,
                        unfocusedContainerColor = Color.Transparent,
                        cursorColor = AppColors.Brand,
                    ),
                )
                Text(
                    stringResource(R.string.text_counter, state.content.length, ReportViewModel.MAX_CONTENT),
                    fontSize = AppType.Caption,
                    color = AppColors.TextTertiary,
                    modifier = Modifier
                        .align(Alignment.End)
                        .padding(top = AppSpacing.Xs),
                )

                Spacer(Modifier.height(AppSpacing.L))
                Text(
                    stringResource(R.string.report_images_label, ReportViewModel.MAX_IMAGES),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                )
                ImageUploadGrid(
                    items = state.images.map {
                        ImageUploadItem(it.id, it.url, it.localUri, it.uploading, it.failed)
                    },
                    canAdd = state.images.size < ReportViewModel.MAX_IMAGES,
                    onAdd = {
                        pickImages.launch(
                            PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly),
                        )
                    },
                    onRemove = viewModel::removeImage,
                    onRetry = viewModel::retryImage,
                    modifier = Modifier.padding(top = InnerGap),
                )
            }

            Text(
                stringResource(R.string.report_notice),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = AppSpacing.M),
            )

            AppPrimaryButton(
                text = stringResource(
                    when {
                        state.loading -> R.string.common_submitting
                        state.hasUploadingImages -> R.string.image_uploading
                        else -> R.string.report_submit
                    },
                ),
                onClick = { viewModel.submit(targetType, targetId) },
                enabled = !state.loading && !state.hasUploadingImages && targetId > 0,
                modifier = Modifier.padding(top = AppSpacing.Xl),
            )
        }
    }
}

@Composable
private fun ReportCard(content: @Composable ColumnScope.() -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppShapes.Card)
            .background(AppColors.CardBg)
            .padding(CardPadding),
        content = content,
    )
}
