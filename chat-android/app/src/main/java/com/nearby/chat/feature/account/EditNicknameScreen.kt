package com.nearby.chat.feature.account

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.imeOnTap

private val DisabledSaveBg = AppColors.Brand.copy(alpha = 0.35f)
private val SaveButtonMinWidth = 40.dp
private val NicknameFieldHeight = 44.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditNicknameScreen(
    onBack: () -> Unit,
    viewModel: EditNicknameViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val latestOnBack by rememberUpdatedState(onBack)
    LaunchedEffect(Unit) { viewModel.load() }
    LaunchedEffect(viewModel) {
        viewModel.saved.collect { latestOnBack() }
    }

    val focusRequester = remember { FocusRequester() }
    LaunchedEffect(Unit) { focusRequester.requestFocus() }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.nickname_edit_title)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    Text(
                        text = stringResource(if (state.saving) R.string.common_saving_short else R.string.common_save),
                        color = Color.White,
                        fontSize = AppType.Subhead,
                        textAlign = TextAlign.Center,
                        modifier = Modifier
                            .padding(end = AppSpacing.M)
                            .clip(AppShapes.Tag)
                            .background(if (state.canSave) AppColors.Brand else DisabledSaveBg)
                            .clickable(
                                enabled = state.canSave,
                                indication = null,
                                interactionSource = remember { MutableInteractionSource() },
                                onClick = viewModel::save,
                            )
                            .padding(horizontal = AppSpacing.M, vertical = 6.dp)
                            .widthIn(min = SaveButtonMinWidth),
                    )
                },
                colors = appTopAppBarColors(AppColors.CardBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .padding(horizontal = AppSpacing.Xl)
                .padding(top = 28.dp),
        ) {
            BasicTextField(
                value = state.draft,
                onValueChange = viewModel::onDraft,
                singleLine = true,
                textStyle = TextStyle(
                    fontSize = AppType.PageTitle,
                    color = AppColors.TextPrimary,
                    lineHeight = AppType.LineBody,
                ),
                cursorBrush = SolidColor(AppColors.Brand),
                modifier = Modifier
                    .fillMaxWidth()
                    .imeOnTap()
                    .focusRequester(focusRequester),
                decorationBox = { inner ->
                    Column(modifier = Modifier.fillMaxWidth()) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(NicknameFieldHeight),
                            contentAlignment = Alignment.CenterStart,
                        ) {
                            if (state.draft.isEmpty()) {
                                Text(
                                    stringResource(R.string.nickname_hint),
                                    color = AppColors.TextTertiary,
                                    fontSize = AppType.PageTitle,
                                )
                            }
                            inner()
                        }
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(1.dp)
                                .background(AppColors.BorderSoft),
                        )
                    }
                },
            )
        }
    }
}
