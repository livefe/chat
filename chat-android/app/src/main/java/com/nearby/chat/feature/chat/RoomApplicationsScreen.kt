package com.nearby.chat.feature.chat

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.RoomApplicationItem
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.appTopAppBarColors

private val StatusPendingColor = AppColors.ReviewPending
private val StatusEnabledColor = AppColors.ReviewApproved
private val StatusRejectedColor = AppColors.ReviewRejected
private val ApplicationRowVPad = 14.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RoomApplicationsScreen(
    onBack: () -> Unit,
    onOpenRoom: (Long) -> Unit,
    viewModel: RoomApplicationsViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.room_create_applications)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        AppPullToRefreshBox(
            isRefreshing = state.loading && state.list.isNotEmpty(),
            onRefresh = viewModel::refresh,
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            if (state.loading && state.list.isEmpty()) {
                CircularProgressIndicator(
                    color = AppColors.Brand,
                    strokeWidth = 2.dp,
                    modifier = Modifier.align(Alignment.Center),
                )
            } else if (state.list.isEmpty()) {
                EmptyHint(stringResource(R.string.room_applications_empty), Modifier.fillMaxSize())
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = AppSpacing.L),
                ) {
                    itemsIndexed(state.list, key = { _, item -> item.id }) { index, item ->
                        if (index > 0) {
                            HorizontalDivider(color = AppColors.Border)
                        }
                        ApplicationRow(
                            item = item,
                            onClick = { viewModel.openableRoomId(item)?.let(onOpenRoom) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ApplicationRow(
    item: RoomApplicationItem,
    onClick: () -> Unit,
) {
    val statusColor = when (item.status) {
        RoomApplicationsViewModel.STATUS_PENDING -> StatusPendingColor
        RoomApplicationsViewModel.STATUS_ENABLED -> StatusEnabledColor
        RoomApplicationsViewModel.STATUS_REJECTED -> StatusRejectedColor
        else -> AppColors.TextTertiary
    }
    val typeLabel = item.placeTypeText.ifBlank { stringResource(RoomPlaceType.label(item.placeType)) }
    val separator = stringResource(R.string.meta_separator)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = ApplicationRowVPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                item.name,
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                listOfNotNull(typeLabel, item.createdAt.ifBlank { null }).joinToString(separator),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = AppSpacing.Xs),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        Text(
            item.statusText,
            fontSize = AppType.Caption,
            fontWeight = FontWeight.Medium,
            color = statusColor,
            modifier = Modifier.padding(start = AppSpacing.M),
        )
    }
}
