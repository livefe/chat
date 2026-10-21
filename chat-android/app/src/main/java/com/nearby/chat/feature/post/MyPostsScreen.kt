package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.ScanViewfinderIcon
import com.nearby.chat.core.ui.appTopAppBarColors

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MyPostsScreen(
    onBack: () -> Unit,
    onOpenPost: (Long) -> Unit,
    onEditPost: (Long) -> Unit = {},
    onOpenSignupList: (Long) -> Unit = {},
    onOpenCheckInCode: (Long) -> Unit = {},
    onOpenScan: () -> Unit = {},
    refreshTick: Boolean = false,
    onRefreshConsumed: () -> Unit = {},
    viewModel: MyPostsViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var confirmTarget by remember { mutableStateOf<MyPostConfirm?>(null) }
    val joined = state.mode == MyPostsMode.Joined

    LaunchedEffect(refreshTick) {
        if (refreshTick) {
            viewModel.refresh()
            onRefreshConsumed()
        }
    }

    InfiniteScrollEffect(
        listState = listState,
        enabled = state.hasMore && !state.loading,
        onNearEnd = viewModel::loadMore,
    )

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = {
                    AppTopBarTitle(stringResource(if (joined) R.string.my_posts_joined else R.string.my_posts_published))
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    if (!joined) {
                        IconButton(onClick = onOpenScan) {
                            Icon(
                                imageVector = ScanViewfinderIcon,
                                contentDescription = stringResource(R.string.scan_action),
                                tint = AppColors.TextPrimary,
                                modifier = Modifier.size(AppIconSize.Nav),
                            )
                        }
                    }
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        AppPullToRefreshBox(
            isRefreshing = state.refreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            when {
                state.loading && state.list.isEmpty() && !state.refreshing -> {
                    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                    }
                }

                !state.hasMore && state.list.isEmpty() -> {
                    EmptyHint(
                        stringResource(if (joined) R.string.my_posts_empty_joined else R.string.my_posts_empty_published),
                        Modifier.fillMaxSize(),
                    )
                }

                else -> {
                    LazyColumn(
                        state = listState,
                        modifier = Modifier
                            .fillMaxSize()
                            .background(AppColors.CardBg),
                    ) {
                        items(state.list, key = { it.id }) { post ->
                            MyPostRow(
                                post = post,
                                mode = state.mode,
                                actingId = state.actingId,
                                onOpenPost = onOpenPost,
                                onEditPost = onEditPost,
                                onOpenSignupList = onOpenSignupList,
                                onOpenCheckInCode = onOpenCheckInCode,
                                onConfirm = { confirmTarget = it },
                            )
                        }
                        if (state.loading && state.list.isNotEmpty() && !state.refreshing) {
                            item {
                                Box(
                                    Modifier
                                        .fillMaxWidth()
                                        .padding(AppSpacing.L),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    CircularProgressIndicator(
                                        color = AppColors.Brand,
                                        strokeWidth = 2.dp,
                                        modifier = Modifier.padding(AppSpacing.Xs),
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    confirmTarget?.let { target ->
        val (title, message) = when (target) {
            is MyPostConfirm.Withdraw -> R.string.post_withdraw to R.string.post_withdraw_confirm
            is MyPostConfirm.DeletePublished -> R.string.common_delete to R.string.post_delete_published_confirm
            is MyPostConfirm.DeleteJoined -> R.string.post_remove_joined to R.string.post_remove_joined_confirm
        }
        AppConfirmDialog(
            title = stringResource(title),
            message = stringResource(message),
            danger = true,
            onConfirm = {
                confirmTarget = null
                when (target) {
                    is MyPostConfirm.Withdraw -> viewModel.withdraw(target.post)
                    is MyPostConfirm.DeletePublished -> viewModel.deletePublished(target.post)
                    is MyPostConfirm.DeleteJoined -> viewModel.deleteJoined(target.post)
                }
            },
            onDismiss = { confirmTarget = null },
        )
    }
}
