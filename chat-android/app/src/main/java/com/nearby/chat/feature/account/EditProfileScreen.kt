package com.nearby.chat.feature.account

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.appTopAppBarColors

private val ProfileAvatarSize = 40.dp
private val ProfileRowHeight = 56.dp
private val ProfileLabelWidth = 72.dp
private val HairlineThickness = 0.5.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditProfileScreen(
    onBack: () -> Unit,
    onEditNickname: () -> Unit,
    onEditMobile: () -> Unit,
    viewModel: EditProfileViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) viewModel.load()
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    val pickAvatar = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickVisualMedia(),
    ) { uri ->
        if (uri != null) viewModel.pickAvatar(uri)
    }
    val notSet = stringResource(R.string.profile_not_set)

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.profile_edit_title)) },
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
                .padding(top = AppSpacing.M),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(AppColors.CardBg),
            ) {
                ProfileNavRow(
                    title = stringResource(R.string.profile_avatar),
                    enabled = !state.uploadingAvatar && !state.saving,
                    onClick = {
                        pickAvatar.launch(
                            PickVisualMediaRequest(
                                ActivityResultContracts.PickVisualMedia.ImageOnly,
                            ),
                        )
                    },
                    trailing = {
                        UserAvatar(
                            url = state.localAvatarUri.ifBlank { state.avatarUrl },
                            size = ProfileAvatarSize,
                        )
                    },
                )
                ProfileDivider()
                ProfileNavRow(
                    title = stringResource(R.string.profile_nickname),
                    value = state.nickname.ifBlank { notSet },
                    enabled = !state.saving,
                    onClick = onEditNickname,
                )
                ProfileDivider()
                ProfileNavRow(
                    title = stringResource(R.string.auth_mobile_label),
                    value = state.mobile.ifBlank { notSet },
                    enabled = !state.saving,
                    onClick = onEditMobile,
                )
            }
        }
    }
}

@Composable
private fun ProfileDivider() {
    HorizontalDivider(
        color = AppColors.BorderSoft,
        thickness = HairlineThickness,
        modifier = Modifier.padding(start = AppSpacing.L),
    )
}

@Composable
private fun ProfileNavRow(
    title: String,
    onClick: () -> Unit,
    enabled: Boolean = true,
    value: String? = null,
    trailing: (@Composable () -> Unit)? = null,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .fillMaxWidth()
            .height(ProfileRowHeight)
            .clickable(
                enabled = enabled,
                indication = null,
                interactionSource = remember { MutableInteractionSource() },
                onClick = onClick,
            )
            .padding(horizontal = AppSpacing.L),
    ) {
        Text(
            title,
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            modifier = Modifier.width(ProfileLabelWidth),
        )
        Box(
            modifier = Modifier.weight(1f),
            contentAlignment = Alignment.CenterEnd,
        ) {
            when {
                trailing != null -> trailing()

                value != null -> Text(
                    value,
                    fontSize = AppType.Body,
                    color = AppColors.TextPrimary,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    textAlign = TextAlign.End,
                )
            }
        }
        Icon(
            imageVector = Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
            modifier = Modifier
                .padding(start = AppSpacing.Xs)
                .size(AppIconSize.Cell),
        )
    }
}
