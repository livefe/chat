package com.nearby.chat.feature.main

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.ime
import androidx.compose.foundation.layout.navigationBars
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ChatBubbleOutline
import androidx.compose.material.icons.outlined.Explore
import androidx.compose.material.icons.outlined.LocationOn
import androidx.compose.material.icons.outlined.PersonOutline
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalTextToolbar
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppUnreadBadge
import com.nearby.chat.feature.chat.NearbyChatScreen
import com.nearby.chat.feature.chat.NearbyChatViewModel
import com.nearby.chat.feature.discover.DiscoverScreen
import com.nearby.chat.feature.me.MeScreen
import com.nearby.chat.feature.messages.MessagesScreen

private const val TAB_NEARBY = 0
private const val TAB_DISCOVER = 1
private const val TAB_MESSAGES = 2
private const val TAB_ME = 3
private val TabBarHeight = 56.dp
private val TabIconSize = 26.dp
private val TabBadgeSize = 16.dp
private val TabBadgeFontSize = 9.sp
private val TabItemTopPadding = 6.dp
private val TabItemBottomPadding = 5.dp

@Composable
fun MainScaffold(
    onOpenPrivateChat: (Long, String) -> Unit,
    onOpenSupport: () -> Unit,
    onOpenPostDetail: (Long) -> Unit,
    onOpenRoom: (Long) -> Unit,
    onOpenNearbyRooms: () -> Unit,
    onOpenEditProfile: () -> Unit,
    onOpenMyPublished: () -> Unit,
    onOpenMyJoined: () -> Unit,
    onOpenFeedback: () -> Unit,
    onOpenSettings: () -> Unit,
    onOpenUserProfile: (Long) -> Unit,
    onOpenPostCreate: (lat: Double?, lng: Double?, placeName: String?) -> Unit,
    onOpenLocationPick: () -> Unit,
    onLogout: () -> Unit,
    chatViewModel: NearbyChatViewModel,
    onOpenShareLocation: () -> Unit = {},
    onReportMessage: (MessageItem) -> Unit = {},
    onOpenForward: (MessageItem) -> Unit = {},
    locationPickLat: Double? = null,
    locationPickLng: Double? = null,
    locationPickPlace: String? = null,
    onLocationPickConsumed: () -> Unit = {},
    viewModel: MainScaffoldViewModel = hiltViewModel(),
) {
    var tab by rememberSaveable { mutableIntStateOf(TAB_NEARBY) }
    val unread by viewModel.totalUnread.collectAsStateWithLifecycle()

    val lifecycleOwner = LocalLifecycleOwner.current
    var mainStarted by remember {
        mutableStateOf(lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED))
    }
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            mainStarted = lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED)
            if (event == Lifecycle.Event.ON_RESUME) {
                viewModel.refreshUnread()
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    LaunchedEffect(locationPickLat, locationPickLng, locationPickPlace) {
        val lat = locationPickLat
        val lng = locationPickLng
        if (lat != null && lng != null) {
            chatViewModel.applyPickedLocation(lat, lng, locationPickPlace.orEmpty())
            onLocationPickConsumed()
        }
    }

    val density = LocalDensity.current
    val textToolbar = LocalTextToolbar.current
    val imeBottomPx = WindowInsets.ime.getBottom(density)
    val navBottomPx = WindowInsets.navigationBars.getBottom(density)
    val tabBarTotalPx = with(density) { TabBarHeight.roundToPx() } + navBottomPx
    val bottomInsetPx = maxOf(tabBarTotalPx, imeBottomPx)

    LaunchedEffect(tab) {
        textToolbar.hide()
    }
    LaunchedEffect(tab, mainStarted) {
        val active = tab == TAB_NEARBY && mainStarted
        chatViewModel.setNearbyPresencePolling(active)
        if (active) chatViewModel.onNearbyTabVisible()
    }

    Column(modifier = Modifier.fillMaxSize()) {
        Box(Modifier.weight(1f).fillMaxWidth()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .alpha(if (tab == TAB_NEARBY) 1f else 0f)
                    .zIndex(if (tab == TAB_NEARBY) 1f else 0f),
            ) {
                NearbyChatScreen(
                    onOpenRoom = onOpenRoom,
                    onOpenNearbyRooms = onOpenNearbyRooms,
                    onOpenUser = onOpenUserProfile,
                    onOpenLocationPick = onOpenLocationPick,
                    onOpenShareLocation = onOpenShareLocation,
                    onOpenPost = onOpenPostDetail,
                    onReportMessage = onReportMessage,
                    onOpenForward = onOpenForward,
                    viewModel = chatViewModel,
                )
            }
            when (tab) {
                TAB_DISCOVER -> DiscoverScreen(
                    onOpenPost = onOpenPostDetail,
                    onOpenRoom = onOpenRoom,
                    onCreatePost = {
                        val s = chatViewModel.uiState.value
                        if (s.locationReady) {
                            onOpenPostCreate(s.lat, s.lng, s.placeName)
                        } else {
                            onOpenPostCreate(null, null, null)
                        }
                    },
                )

                TAB_MESSAGES -> MessagesScreen(onOpenChat = onOpenPrivateChat)

                TAB_ME -> MeScreen(
                    onEditProfile = onOpenEditProfile,
                    onMyPublished = onOpenMyPublished,
                    onMyJoined = onOpenMyJoined,
                    onSupport = onOpenSupport,
                    onFeedback = onOpenFeedback,
                    onSettings = onOpenSettings,
                    onLogout = onLogout,
                )
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(with(density) { bottomInsetPx.toDp() }),
        ) {
            val tabAlpha = (1f - imeBottomPx.toFloat() / tabBarTotalPx.coerceAtLeast(1)).coerceIn(0f, 1f)
            AppTabBar(
                selected = tab,
                unread = unread,
                onSelect = { tab = it },
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .alpha(tabAlpha),
            )
        }
    }
}

@Composable
private fun AppTabBar(
    selected: Int,
    unread: Int,
    onSelect: (Int) -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(AppColors.CardBg)
            .drawBehind {
                drawLine(
                    color = AppColors.Border,
                    start = Offset(0f, 0f),
                    end = Offset(size.width, 0f),
                    strokeWidth = 1.dp.toPx(),
                )
            }
            .windowInsetsPadding(WindowInsets.navigationBars),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(TabBarHeight),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TabItem(
                selected = selected == TAB_NEARBY,
                icon = Icons.Outlined.LocationOn,
                label = stringResource(R.string.tab_nearby),
                onClick = { onSelect(TAB_NEARBY) },
            )
            TabItem(
                selected = selected == TAB_DISCOVER,
                icon = Icons.Outlined.Explore,
                label = stringResource(R.string.tab_discover),
                onClick = { onSelect(TAB_DISCOVER) },
            )
            TabItem(
                selected = selected == TAB_MESSAGES,
                icon = Icons.Outlined.ChatBubbleOutline,
                label = stringResource(R.string.tab_messages),
                badge = unread,
                onClick = { onSelect(TAB_MESSAGES) },
            )
            TabItem(
                selected = selected == TAB_ME,
                icon = Icons.Outlined.PersonOutline,
                label = stringResource(R.string.tab_me),
                onClick = { onSelect(TAB_ME) },
            )
        }
    }
}

@Composable
private fun RowScope.TabItem(
    selected: Boolean,
    icon: ImageVector,
    label: String,
    badge: Int = 0,
    onClick: () -> Unit,
) {
    val color = if (selected) AppColors.Brand else AppColors.TextPrimary
    Column(
        modifier = Modifier
            .weight(1f)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(top = TabItemTopPadding, bottom = TabItemBottomPadding),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Box(contentAlignment = Alignment.TopEnd) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                tint = color,
                modifier = Modifier.size(TabIconSize),
            )
            if (badge > 0) {
                AppUnreadBadge(
                    count = badge,
                    size = TabBadgeSize,
                    fontSize = TabBadgeFontSize,
                    modifier = Modifier.offset(x = 6.dp, y = (-3).dp),
                )
            }
        }
        Spacer(Modifier.height(AppSpacing.Xxs))
        Text(
            text = label,
            color = color,
            fontSize = AppType.Caption,
            lineHeight = AppType.Caption,
            fontWeight = if (selected) FontWeight.Medium else FontWeight.Normal,
        )
    }
}
