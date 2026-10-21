package com.nearby.chat.feature.account

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.BlockListItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.appTopAppBarColors

private val BlockAvatarSize = 40.dp
private val BlockDividerInset = AppSpacing.L + BlockAvatarSize + AppSpacing.M

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BlockListScreen(
    onBack: () -> Unit,
    onOpenProfile: (Long, String, String) -> Unit,
    viewModel: BlockListViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var pendingUnblock by remember { mutableStateOf<BlockListItem?>(null) }

    InfiniteScrollEffect(
        listState = listState,
        enabled = state.hasMore && !state.loading,
        onNearEnd = viewModel::loadMore,
    )

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.block_list_title)) },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            when {
                state.loading && state.list.isEmpty() -> {
                    CircularProgressIndicator(
                        color = AppColors.Brand,
                        strokeWidth = 2.dp,
                        modifier = Modifier.align(Alignment.Center),
                    )
                }

                !state.hasMore && state.list.isEmpty() -> {
                    EmptyHint(stringResource(R.string.block_list_empty), Modifier.fillMaxSize())
                }

                else -> {
                    LazyColumn(
                        state = listState,
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M)
                            .clip(AppShapes.Card)
                            .background(AppColors.CardBg),
                    ) {
                        items(state.list, key = { it.userId }) { item ->
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable {
                                        onOpenProfile(item.userId, item.nickname, item.avatarUrl)
                                    }
                                    .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
                            ) {
                                UserAvatar(url = item.avatarUrl, size = BlockAvatarSize)
                                Text(
                                    item.nickname,
                                    fontSize = AppType.Body,
                                    fontWeight = FontWeight.Medium,
                                    color = AppColors.TextPrimary,
                                    modifier = Modifier
                                        .weight(1f)
                                        .padding(horizontal = AppSpacing.M),
                                )
                                TextButton(
                                    onClick = { pendingUnblock = item },
                                    enabled = state.unblockingId == 0L,
                                ) {
                                    Text(stringResource(R.string.block_unblock), color = AppColors.Brand)
                                }
                            }
                            HorizontalDivider(
                                color = AppColors.Divider,
                                modifier = Modifier.padding(start = BlockDividerInset),
                            )
                        }
                        if (state.loading && state.list.isNotEmpty()) {
                            item {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(AppSpacing.L),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    CircularProgressIndicator(
                                        color = AppColors.Brand,
                                        strokeWidth = 2.dp,
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    pendingUnblock?.let { item ->
        AppConfirmDialog(
            title = stringResource(R.string.block_unblock_title),
            message = stringResource(R.string.block_unblock_confirm, item.nickname),
            onConfirm = {
                pendingUnblock = null
                viewModel.unblock(item)
            },
            onDismiss = { pendingUnblock = null },
        )
    }
}
