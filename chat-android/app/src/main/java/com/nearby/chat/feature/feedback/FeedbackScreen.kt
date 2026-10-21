package com.nearby.chat.feature.feedback

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
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
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.imeOnTap

private val FeedbackContentHeight = 140.dp

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun FeedbackScreen(
    onBack: () -> Unit,
    onOpenRecords: () -> Unit,
    viewModel: FeedbackViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnBack by rememberUpdatedState(onBack)
    val pickImages = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickMultipleVisualMedia(FeedbackViewModel.MAX_IMAGES),
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
                title = { AppTopBarTitle(stringResource(R.string.me_feedback)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    TextButton(
                        onClick = onOpenRecords,
                        colors = ButtonDefaults.textButtonColors(contentColor = AppColors.TextSecondary),
                    ) {
                        Text(stringResource(R.string.feedback_records), fontSize = AppType.Secondary)
                    }
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
                .padding(bottom = AppSpacing.Xxl),
        ) {
            Column(
                modifier = Modifier
                    .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M)
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg)
                    .padding(AppSpacing.L),
            ) {
                SectionLabel(stringResource(R.string.feedback_type_label))
                FlowRow(
                    horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    verticalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    modifier = Modifier.padding(top = 10.dp),
                ) {
                    FeedbackType.entries.forEach { type ->
                        AppFilterChip(
                            label = stringResource(type.label),
                            selected = state.type == type.value,
                            onClick = { viewModel.onType(type.value) },
                        )
                    }
                }

                Spacer(Modifier.height(AppSpacing.L))
                SectionLabel(stringResource(R.string.feedback_content_label))
                OutlinedTextField(
                    value = state.content,
                    onValueChange = viewModel::onContent,
                    placeholder = {
                        Text(stringResource(R.string.feedback_content_hint), color = AppColors.TextTertiary)
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .imeOnTap()
                        .padding(top = AppSpacing.S)
                        .height(FeedbackContentHeight),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = AppColors.Brand,
                        unfocusedBorderColor = AppColors.BorderSoft,
                    ),
                )
                Text(
                    stringResource(R.string.text_counter, state.content.length, FeedbackViewModel.MAX_CONTENT),
                    fontSize = AppType.Caption,
                    color = AppColors.TextTertiary,
                    modifier = Modifier
                        .align(Alignment.End)
                        .padding(top = AppSpacing.Xs),
                )

                Spacer(Modifier.height(AppSpacing.M))
                SectionLabel(stringResource(R.string.feedback_contact_label))
                OutlinedTextField(
                    value = state.contact,
                    onValueChange = viewModel::onContact,
                    placeholder = {
                        Text(stringResource(R.string.feedback_contact_hint), color = AppColors.TextTertiary)
                    },
                    singleLine = true,
                    modifier = Modifier
                        .fillMaxWidth()
                        .imeOnTap()
                        .padding(top = AppSpacing.S),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = AppColors.Brand,
                        unfocusedBorderColor = AppColors.BorderSoft,
                    ),
                )
            }

            Column(
                modifier = Modifier
                    .padding(horizontal = AppSpacing.L)
                    .fillMaxWidth()
                    .clip(AppShapes.Card)
                    .background(AppColors.CardBg)
                    .padding(AppSpacing.L),
            ) {
                SectionLabel(stringResource(R.string.feedback_images_label, FeedbackViewModel.MAX_IMAGES))
                ImageUploadGrid(
                    items = state.images.map {
                        ImageUploadItem(it.id, it.url, it.localUri, it.uploading, it.failed)
                    },
                    canAdd = state.images.size < FeedbackViewModel.MAX_IMAGES,
                    onAdd = {
                        pickImages.launch(
                            PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly),
                        )
                    },
                    onRemove = viewModel::removeImage,
                    onRetry = viewModel::retryImage,
                    modifier = Modifier.padding(top = 10.dp),
                )
            }

            AppPrimaryButton(
                text = stringResource(
                    when {
                        state.loading -> R.string.common_submitting
                        state.hasUploadingImages -> R.string.image_uploading
                        else -> R.string.feedback_submit
                    },
                ),
                onClick = viewModel::submit,
                enabled = !state.loading && !state.hasUploadingImages,
                modifier = Modifier.padding(horizontal = AppSpacing.L, vertical = AppSpacing.Xxl),
            )
        }
    }
}

@Composable
private fun SectionLabel(text: String) {
    Text(text, fontSize = AppType.Secondary, color = AppColors.TextSecondary)
}
