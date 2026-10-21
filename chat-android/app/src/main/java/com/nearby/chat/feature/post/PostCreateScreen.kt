package com.nearby.chat.feature.post

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppSheetTitle
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.DateTimeWheelSheet
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.asString
import com.nearby.chat.feature.post.ui.CategoryPickGrid
import com.nearby.chat.feature.post.ui.rememberDismissInput

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostCreateScreen(
    onBack: () -> Unit,
    onCreated: (Long) -> Unit,
    onPickLocation: () -> Unit,
    roomId: Long = 0,
    postId: Long = 0,
    viewModel: PostCreateViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val dismissInput = rememberDismissInput()
    val isEdit = postId > 0 || state.isEdit
    val inForm = isEdit || state.inForm
    var categorySheet by remember { mutableStateOf(false) }
    var dateTimeTarget by remember { mutableStateOf<PostDateTimeTarget?>(null) }
    val currentOnCreated by rememberUpdatedState(onCreated)
    val currentOnBack by rememberUpdatedState(onBack)

    LaunchedEffect(viewModel) {
        viewModel.submitted.collect { currentOnCreated(it) }
    }
    LaunchedEffect(viewModel) {
        viewModel.unavailable.collect { currentOnBack() }
    }
    LaunchedEffect(roomId) {
        if (postId <= 0) viewModel.bindRoom(roomId)
    }
    LaunchedEffect(postId) {
        if (postId > 0) viewModel.loadForEdit(postId)
    }
    LaunchedEffect(inForm, isEdit) {
        if (!isEdit && inForm) viewModel.ensureLocation()
    }
    LaunchedEffect(state.locationReady, state.latitude, state.longitude) {
        if (state.locationReady) viewModel.refreshNearbyRooms()
    }

    val pageTitle = stringResource(
        when {
            isEdit -> R.string.post_edit_title
            !inForm -> R.string.post_pick_title
            else -> R.string.post_publish
        },
    )
    val locationLabel = when {
        state.locating -> stringResource(R.string.post_locating)

        state.locationReady -> PlaceNames.display(
            state.placeName,
            isCurrentGps = true,
            nearbyLabel = stringResource(R.string.place_nearby),
        )

        else -> state.locationError?.asString() ?: stringResource(R.string.post_location_pick_hint)
    }
    val canSubmit = !state.loading && !state.pageLoading && !state.hasUploadingImages && state.locationReady

    val pickImages = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickMultipleVisualMedia(PostCreateUiState.MAX_IMAGES),
    ) { uris ->
        if (uris.isNotEmpty()) viewModel.addImages(uris)
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(pageTitle) },
                navigationIcon = {
                    AppBackIconButton(
                        onClick = {
                            if (!isEdit && inForm) viewModel.backToTypePick() else onBack()
                        },
                    )
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        when {
            state.pageLoading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                }
            }

            !inForm -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .verticalScroll(rememberScrollState())
                        .padding(horizontal = AppSpacing.L, vertical = AppSpacing.Xl),
                ) {
                    Text(
                        stringResource(R.string.post_pick_hint),
                        fontSize = AppType.Secondary,
                        color = AppColors.TextPrimary,
                        modifier = Modifier.padding(bottom = AppSpacing.L),
                    )
                    CategoryPickGrid(
                        selected = state.category,
                        onSelect = viewModel::onCategory,
                    )
                }
            }

            else -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                ) {
                    PostCreateFormContent(
                        state = state,
                        viewModel = viewModel,
                        isEdit = isEdit,
                        locationLabel = locationLabel,
                        onDismissInput = dismissInput,
                        onOpenCategorySheet = { categorySheet = true },
                        onPickLocation = onPickLocation,
                        onPickDateTime = { dateTimeTarget = it },
                        onPickImages = {
                            if (state.images.size < PostCreateUiState.MAX_IMAGES) {
                                pickImages.launch(
                                    PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly),
                                )
                            }
                        },
                        modifier = Modifier
                            .weight(1f)
                            .verticalScroll(rememberScrollState())
                            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.S),
                    )
                    AppPrimaryButton(
                        text = stringResource(
                            when {
                                state.loading && isEdit -> R.string.post_saving
                                state.loading -> R.string.post_submitting
                                isEdit -> R.string.post_save_submit
                                else -> R.string.post_publish
                            },
                        ),
                        onClick = viewModel::submit,
                        enabled = canSubmit,
                        modifier = Modifier.padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
                    )
                    Spacer(Modifier.height(AppSpacing.S))
                }
            }
        }
    }

    if (categorySheet && !isEdit) {
        AppModalBottomSheet(
            onDismissRequest = { categorySheet = false },
        ) {
            AppSheetTitle(stringResource(R.string.post_change_category))
            CategoryPickGrid(
                selected = state.category,
                onSelect = { cat ->
                    viewModel.onCategory(cat)
                    categorySheet = false
                },
                modifier = Modifier.padding(horizontal = AppSpacing.L, vertical = AppSpacing.S),
            )
        }
    }

    dateTimeTarget?.let { target ->
        val pickerInitial = remember(target) {
            when (target) {
                PostDateTimeTarget.Start -> state.startAt
                PostDateTimeTarget.End -> state.endAt.takeIf { it > 0 } ?: state.startAt
                PostDateTimeTarget.SignupDeadline -> state.signupDeadlineAt.takeIf { it > 0 } ?: state.startAt
            }
        }
        DateTimeWheelSheet(
            visible = true,
            title = stringResource(
                when (target) {
                    PostDateTimeTarget.Start -> R.string.post_pick_start
                    PostDateTimeTarget.End -> R.string.post_pick_end
                    PostDateTimeTarget.SignupDeadline -> R.string.post_pick_deadline
                },
            ),
            initialSeconds = pickerInitial,
            onChange = { seconds ->
                when (target) {
                    PostDateTimeTarget.Start -> viewModel.onStartAt(seconds)
                    PostDateTimeTarget.End -> viewModel.onEndAt(seconds)
                    PostDateTimeTarget.SignupDeadline -> viewModel.onSignupDeadlineAt(seconds)
                }
            },
            onDismiss = { dateTimeTarget = null },
            onConfirm = { dateTimeTarget = null },
        )
    }
}
