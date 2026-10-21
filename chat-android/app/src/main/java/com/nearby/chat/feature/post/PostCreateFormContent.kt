package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.post.PostCategories
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.ImageUploadGrid
import com.nearby.chat.core.ui.ImageUploadItem
import com.nearby.chat.feature.post.ui.PostClickRow
import com.nearby.chat.feature.post.ui.PostFormField
import com.nearby.chat.feature.post.ui.PostFormLabelWidth
import com.nearby.chat.feature.post.ui.PostSwitchRow

@Composable
internal fun PostCreateFormContent(
    state: PostCreateUiState,
    viewModel: PostCreateViewModel,
    isEdit: Boolean,
    locationLabel: String,
    onDismissInput: () -> Unit,
    onOpenCategorySheet: () -> Unit,
    onPickLocation: () -> Unit,
    onPickDateTime: (PostDateTimeTarget) -> Unit,
    onPickImages: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(modifier = modifier) {
        CategoryHeader(
            category = state.category,
            editable = !isEdit,
            onClick = {
                onDismissInput()
                onOpenCategorySheet()
            },
        )

        PostFormCard {
            PostFormField(
                label = stringResource(R.string.post_field_title),
                value = state.title,
                onValueChange = viewModel::onTitle,
                placeholder = stringResource(R.string.post_field_title_hint),
                singleLine = true,
            )
            if (state.isRoomPost && state.category == PostCategories.NOTICE) {
                PostFormDivider()
                PostSwitchRow(stringResource(R.string.post_pinned), state.pinned, viewModel::onPinned)
            }
            PostFormDivider()
            val showNearby = state.locationReady &&
                (state.displayNearbyRooms.isNotEmpty() || state.nearbyLoading)
            PostClickRow(
                label = stringResource(R.string.post_field_location),
                value = if (state.locationReady) locationLabel else "",
                placeholder = locationLabel,
                compactBottom = showNearby,
                onClick = {
                    onDismissInput()
                    onPickLocation()
                },
            )
            if (showNearby) {
                NearbyRoomSuggestions(
                    rooms = state.displayNearbyRooms,
                    selectedRoomId = state.roomId,
                    loading = state.nearbyLoading,
                    onSelect = { room ->
                        onDismissInput()
                        viewModel.selectAttachRoom(room)
                    },
                )
            }
        }

        PostFormCard {
            PostFormField(
                label = stringResource(R.string.post_field_content),
                value = state.content,
                onValueChange = viewModel::onContent,
                placeholder = stringResource(R.string.post_field_content_hint),
                singleLine = false,
                minLines = 4,
            )
        }

        when (state.category) {
            PostCategories.EVENT -> PostEventSections(
                state = state,
                viewModel = viewModel,
                onPickDateTime = onPickDateTime,
            )

            PostCategories.IDLE -> PostIdleSections(state, viewModel)

            PostCategories.HELP -> PostHelpSections(state, viewModel)
        }

        PostFormCard {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = AppSpacing.L, vertical = 14.dp),
                verticalAlignment = Alignment.Top,
            ) {
                Text(
                    stringResource(R.string.post_field_images),
                    fontSize = AppType.Body,
                    color = AppColors.TextPrimary,
                    modifier = Modifier
                        .width(PostFormShortLabelWidth)
                        .padding(top = AppSpacing.Xxs),
                )
                ImageUploadGrid(
                    items = state.images.map {
                        ImageUploadItem(it.id, it.url, it.localUri, it.uploading, it.failed)
                    },
                    canAdd = state.images.size < PostCreateUiState.MAX_IMAGES,
                    onAdd = onPickImages,
                    onRemove = viewModel::removeImage,
                    onRetry = viewModel::retryImage,
                    modifier = Modifier.weight(1f),
                )
            }
            if (state.hasUploadingImages) {
                Text(
                    stringResource(R.string.post_images_uploading),
                    fontSize = AppType.Caption,
                    color = AppColors.TextTertiary,
                    modifier = Modifier.padding(start = PostFormLabelWidth, end = AppSpacing.L, bottom = AppSpacing.M),
                )
            }
        }
    }
}

@Composable
private fun CategoryHeader(
    category: String,
    editable: Boolean,
    onClick: () -> Unit,
) {
    val meta = PostCategories.option(category)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppShapes.CardLg)
            .background(AppColors.CardBg)
            .then(if (editable) Modifier.clickable(onClick = onClick) else Modifier)
            .padding(horizontal = AppSpacing.L, vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(
            modifier = Modifier.weight(1f),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
        ) {
            Text(
                stringResource(meta?.label ?: PostCategories.label(category)),
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.TextPrimary,
            )
            Text(
                meta?.hint?.let { stringResource(it) }.orEmpty(),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f, fill = false),
            )
        }
        if (editable) {
            Icon(
                Icons.AutoMirrored.Outlined.KeyboardArrowRight,
                contentDescription = null,
                tint = AppColors.TextTertiary,
                modifier = Modifier.size(AppIconSize.Inline),
            )
        }
    }
}
