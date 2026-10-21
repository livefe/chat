package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppFilterChip
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.appTopAppBarColors

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostSignupListScreen(
    onBack: () -> Unit,
    viewModel: PostSignupListViewModel = hiltViewModel(),
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
                title = {
                    AppTopBarTitle(stringResource(R.string.signup_list_title))
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    TextButton(
                        onClick = viewModel::export,
                        enabled = !state.exporting,
                    ) {
                        Text(
                            stringResource(if (state.exporting) R.string.signup_exporting else R.string.signup_export_action),
                            color = if (state.exporting) AppColors.TextTertiary else AppColors.TextPrimary,
                            fontSize = AppType.Secondary,
                        )
                    }
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(AppColors.CardBg),
            ) {
                Row(modifier = Modifier.fillMaxWidth()) {
                    SignupTab.statusTabs(state.counts).forEach { (tab, count) ->
                        SignupStatusTab(
                            label = stringResource(tab.label),
                            count = count,
                            selected = state.statusFilter == tab.status,
                            onClick = { viewModel.setStatusFilter(tab.status) },
                            modifier = Modifier.weight(1f),
                        )
                    }
                }
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(1.dp)
                        .background(AppColors.Divider),
                )
                if (state.needCheckIn && state.statusFilter == PostSignupRules.STATUS_CONFIRMED) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .horizontalScroll(rememberScrollState())
                            .padding(horizontal = AppSpacing.M, vertical = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    ) {
                        SignupTab.checkInTabs(state.counts).forEach { (tab, count) ->
                            AppFilterChip(
                                label = stringResource(R.string.signup_filter_count, stringResource(tab.label), count),
                                selected = state.checkInFilter == tab.checkIn,
                                onClick = { tab.checkIn?.let(viewModel::setCheckInFilter) },
                            )
                        }
                    }
                }
            }

            AppPullToRefreshBox(
                isRefreshing = state.refreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.weight(1f),
            ) {
                when {
                    state.loading && state.list.isEmpty() && !state.refreshing -> {
                        Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                            CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                        }
                    }

                    !state.hasMore && state.list.isEmpty() -> {
                        EmptyHint(stringResource(R.string.signup_list_empty))
                    }

                    else -> {
                        LazyColumn(
                            state = listState,
                            modifier = Modifier.fillMaxSize(),
                            contentPadding = PaddingValues(AppSpacing.M),
                            verticalArrangement = Arrangement.spacedBy(10.dp),
                        ) {
                            items(state.list, key = { it.id }) { item ->
                                SignupCard(
                                    item = item,
                                    needCheckIn = state.needCheckIn,
                                    checkInWindowOpen = state.checkInWindowOpen,
                                    checking = state.checkingUserId == item.userId,
                                    onManualCheckIn = { viewModel.manualCheckIn(item.userId) },
                                    onEditRemark = { viewModel.openRemarkEditor(item) },
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
    }

    state.remarkEditing?.let { editing ->
        RemarkEditDialog(
            nickname = editing.userNickname,
            draft = state.remarkDraft,
            maxLength = PostSignupListViewModel.REMARK_MAX,
            saving = state.remarkSaving,
            onDraftChange = viewModel::onRemarkDraftChange,
            onDismiss = viewModel::closeRemarkEditor,
            onConfirm = viewModel::saveRemark,
        )
    }
}
