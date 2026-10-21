package com.nearby.chat.feature.discover

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.items
import androidx.compose.foundation.lazy.staggeredgrid.rememberLazyStaggeredGridState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.external.ExternalLink
import com.nearby.chat.core.model.AdLandingType
import com.nearby.chat.core.model.FeedItem
import com.nearby.chat.core.model.SponsoredItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.feature.post.ui.PostFeedItem
import com.nearby.chat.feature.post.ui.PostFeedVariant
import com.nearby.chat.feature.post.ui.SponsoredFeedCard

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DiscoverScreen(
    onOpenPost: (Long) -> Unit,
    onOpenRoom: (Long) -> Unit,
    onCreatePost: () -> Unit,
    viewModel: DiscoverViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyStaggeredGridState()
    val context = LocalContext.current
    LaunchedEffect(Unit) { viewModel.load() }

    InfiniteScrollEffect(
        gridState = listState,
        enabled = state.hasMore && !state.loading,
        buffer = 4,
        onNearEnd = viewModel::loadMore,
    )

    fun openSponsored(item: SponsoredItem) {
        viewModel.onSponsoredClick(item)
        when (item.landingType) {
            AdLandingType.POST -> {
                item.landingValue.toLongOrNull()?.takeIf { it > 0 }?.let(onOpenPost)
            }

            AdLandingType.ROOM -> {
                item.landingValue.toLongOrNull()?.takeIf { it > 0 }?.let(onOpenRoom)
            }

            else -> {
                ExternalLink.open(context, item.landingValue)
            }
        }
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = {
            TopAppBar(
                title = {
                    AppTopBarTitle(stringResource(R.string.tab_discover))
                },
                actions = {
                    IconButton(onClick = onCreatePost) {
                        Icon(
                            Icons.Outlined.Add,
                            contentDescription = stringResource(R.string.post_publish),
                            tint = AppColors.TextPrimary,
                            modifier = Modifier.size(AppIconSize.Cell),
                        )
                    }
                },
                windowInsets = WindowInsets.statusBars,
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
                    EmptyHint(stringResource(R.string.common_empty_content), Modifier.fillMaxSize())
                }

                else -> {
                    LazyVerticalStaggeredGrid(
                        columns = StaggeredGridCells.Fixed(2),
                        state = listState,
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(AppSpacing.M),
                        verticalItemSpacing = AppSpacing.S,
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        items(
                            state.list,
                            key = { item ->
                                when (item) {
                                    is FeedItem.Post -> "post-${item.post.id}"

                                    is FeedItem.Sponsored ->
                                        "ad-${item.sponsored.requestId}-${item.sponsored.slotIndex}"
                                }
                            },
                        ) { item ->
                            when (item) {
                                is FeedItem.Post -> {
                                    Box(
                                        modifier = Modifier.clickable { onOpenPost(item.post.id) },
                                    ) {
                                        PostFeedItem(
                                            post = item.post,
                                            variant = PostFeedVariant.Waterfall,
                                        )
                                    }
                                }

                                is FeedItem.Sponsored -> {
                                    SponsoredFeedCard(
                                        item = item.sponsored,
                                        onImpression = viewModel::reportSponsoredImpression,
                                        onClick = ::openSponsored,
                                    )
                                }
                            }
                        }
                        if (state.loading && state.list.isNotEmpty() && !state.refreshing) {
                            item(span = StaggeredGridItemSpan.FullLine) {
                                Box(
                                    Modifier
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
}
