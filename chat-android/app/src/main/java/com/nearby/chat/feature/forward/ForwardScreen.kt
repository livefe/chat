package com.nearby.chat.feature.forward

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.ChatBubbleOutline
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.LocalToaster
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.asString

private val NearbyRowVerticalPadding = 14.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ForwardScreen(
    onBack: () -> Unit,
    viewModel: ForwardViewModel = hiltViewModel(),
) {
    val session = viewModel.session
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnBack by rememberUpdatedState(onBack)
    val context = LocalContext.current
    val toaster = LocalToaster.current
    val storagePermission = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (granted) viewModel.saveShareImage() else toaster.fail(R.string.share_save_need_storage)
    }

    fun saveImage() {
        val needsStorage = Build.VERSION.SDK_INT < Build.VERSION_CODES.Q &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.WRITE_EXTERNAL_STORAGE) !=
            PackageManager.PERMISSION_GRANTED
        if (needsStorage) {
            storagePermission.launch(Manifest.permission.WRITE_EXTERNAL_STORAGE)
        } else {
            viewModel.saveShareImage()
        }
    }

    LaunchedEffect(viewModel) {
        if (session == null) {
            latestOnBack()
            return@LaunchedEffect
        }
        viewModel.done.collect { latestOnBack() }
    }

    if (session == null) return

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(session.pageTitle.asString()) },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            NearbyDestinationRow(enabled = !state.sending, onClick = viewModel::sendNearby)

            Spacer(Modifier.height(AppSpacing.S))

            Column(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .background(AppColors.CardBg),
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = AppSpacing.L),
                ) {
                    ForwardTabItem(
                        label = stringResource(R.string.tab_messages),
                        selected = state.tab == ForwardTab.Chats,
                        onClick = { viewModel.selectTab(ForwardTab.Chats) },
                    )
                    Spacer(Modifier.width(AppSpacing.Xl))
                    ForwardTabItem(
                        label = stringResource(R.string.nearby_rooms_title),
                        selected = state.tab == ForwardTab.Rooms,
                        onClick = { viewModel.selectTab(ForwardTab.Rooms) },
                    )
                }

                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth(),
                ) {
                    Column(Modifier.fillMaxSize()) {
                        ForwardSearchField(
                            value = state.query,
                            onValueChange = viewModel::onQuery,
                            placeholder = stringResource(
                                if (state.tab == ForwardTab.Rooms) {
                                    R.string.forward_search_rooms
                                } else {
                                    R.string.forward_search_chats
                                },
                            ),
                            enabled = !state.sending,
                        )
                        when (state.tab) {
                            ForwardTab.Rooms -> ForwardRoomsPane(
                                state = state,
                                rooms = state.filteredRooms,
                                onSelect = { viewModel.sendRoom(it.id) },
                            )

                            ForwardTab.Chats -> ForwardChatsPane(
                                state = state,
                                chats = state.filteredChats,
                                onSelect = { viewModel.sendPrivate(it.peerId) },
                                onLoadMore = viewModel::loadMoreChats,
                            )

                            ForwardTab.Dest -> Unit
                        }
                    }
                }
            }

            if (session.showExternalShare && session.postId > 0L) {
                ForwardExternalShareBar(
                    sharing = state.sharing,
                    onSaveImage = ::saveImage,
                    onCopyLink = viewModel::copyLink,
                    onShare = viewModel::shareExternal,
                )
            }
        }
    }
}

@Composable
private fun NearbyDestinationRow(enabled: Boolean, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg)
            .clickable(enabled = enabled, onClick = onClick)
            .padding(horizontal = AppSpacing.L, vertical = NearbyRowVerticalPadding),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(ForwardLeadingSize)
                .clip(AppShapes.Card)
                .background(AppColors.PageBg),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                Icons.Outlined.ChatBubbleOutline,
                contentDescription = null,
                tint = AppColors.TextSecondary,
                modifier = Modifier.size(AppIconSize.Cell),
            )
        }
        Spacer(Modifier.width(AppSpacing.M))
        Text(
            stringResource(R.string.forward_nearby),
            fontSize = AppType.Subhead,
            fontWeight = FontWeight.SemiBold,
            color = AppColors.TextPrimary,
            modifier = Modifier.weight(1f),
        )
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
        )
    }
}
