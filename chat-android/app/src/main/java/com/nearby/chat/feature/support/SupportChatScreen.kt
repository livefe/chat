package com.nearby.chat.feature.support

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.dismissKeyboardOnTap
import com.nearby.chat.feature.chat.ui.ChatComposerBar
import com.nearby.chat.feature.chat.ui.chatMessageItems
import com.nearby.chat.feature.chat.ui.scrollToLatestIfFollowing

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SupportChatScreen(
    onBack: () -> Unit,
    viewModel: SupportChatViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val displayMessages by viewModel.displayMessages.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var moreOpen by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) { viewModel.enter() }
    DisposableEffect(Unit) {
        onDispose { viewModel.leave() }
    }

    LaunchedEffect(displayMessages.lastOrNull()?.id, displayMessages.lastOrNull()?.clientMsgId) {
        val last = displayMessages.lastOrNull() ?: return@LaunchedEffect
        listState.scrollToLatestIfFollowing(
            last = last,
            selfId = state.selfId,
            reverseLayout = false,
            latestIndex = displayMessages.lastIndex,
        )
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(state.brandName) },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            Modifier
                .padding(padding)
                .fillMaxSize()
                .imePadding(),
        ) {
            val dismissExtras = { moreOpen = false }

            if (state.loading && displayMessages.isEmpty()) {
                EmptyHint(stringResource(R.string.common_loading), Modifier.weight(1f))
            } else if (displayMessages.isEmpty()) {
                EmptyHint(stringResource(R.string.support_empty), Modifier.weight(1f).dismissKeyboardOnTap(dismissExtras))
            } else {
                LazyColumn(
                    state = listState,
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .dismissKeyboardOnTap(dismissExtras),
                    contentPadding = PaddingValues(vertical = AppSpacing.S),
                ) {
                    chatMessageItems(
                        messages = displayMessages,
                        selfId = state.selfId,
                        hideName = true,
                    )
                }
            }

            ChatComposerBar(
                input = state.input,
                onInputChange = viewModel::onInputChange,
                onSend = viewModel::send,
                onImagePicked = viewModel::sendMedia,
                onPickLocation = {},
                onVoiceRecorded = {},
                uploadVoice = { _, _ -> "" },
                sendEnabled = state.input.isNotBlank(),
                uploading = state.sending || state.uploading,
                placeholder = stringResource(R.string.support_placeholder),
                enableVoice = false,
                enableLocation = false,
                onFilePicked = viewModel::sendFile,
                moreOpen = moreOpen,
                onMoreOpenChange = { moreOpen = it },
                modifier = Modifier
                    .background(AppColors.PageBg)
                    .navigationBarsPadding(),
            )
        }
    }
}
