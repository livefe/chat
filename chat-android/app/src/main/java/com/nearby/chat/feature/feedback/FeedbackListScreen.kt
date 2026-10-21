package com.nearby.chat.feature.feedback

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.FeedbackItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.appTopAppBarColors

private val ProgressStroke = 2.dp
private val RowVerticalPadding = 14.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedbackListScreen(
    onBack: () -> Unit,
    onOpenDetail: (Long) -> Unit,
    viewModel: FeedbackListViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    InfiniteScrollEffect(
        listState = listState,
        enabled = state.hasMore && !state.loading,
        onNearEnd = viewModel::loadMore,
    )

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.feedback_records)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
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
                        strokeWidth = ProgressStroke,
                        modifier = Modifier.align(Alignment.Center),
                    )
                }

                !state.hasMore && state.list.isEmpty() -> {
                    EmptyHint(stringResource(R.string.feedback_list_empty), Modifier.fillMaxSize())
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
                        items(state.list, key = { it.id }) { item ->
                            FeedbackListRow(item = item, onClick = { onOpenDetail(item.id) })
                            HorizontalDivider(color = AppColors.Divider, modifier = Modifier.padding(start = AppSpacing.L))
                        }
                        if (state.loading && state.list.isNotEmpty()) {
                            item {
                                Box(
                                    Modifier
                                        .fillMaxWidth()
                                        .padding(AppSpacing.L),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    CircularProgressIndicator(
                                        color = AppColors.Brand,
                                        strokeWidth = ProgressStroke,
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
}

@Composable
private fun FeedbackListRow(item: FeedbackItem, onClick: () -> Unit) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = AppSpacing.L, vertical = RowVerticalPadding),
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                item.typeText.ifBlank { stringResource(R.string.feedback_default_type) },
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
            )
            Text(
                item.content,
                fontSize = AppType.Meta,
                color = AppColors.TextTertiary,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.padding(top = AppSpacing.Xs),
            )
        }
        Text(
            item.statusText,
            fontSize = AppType.Caption,
            color = feedbackStatusColor(item.status),
            modifier = Modifier.padding(start = AppSpacing.S),
        )
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
            modifier = Modifier.padding(start = AppSpacing.Xxs),
        )
    }
}
