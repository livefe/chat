package com.nearby.chat.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RoomCreateScreen(
    onBack: () -> Unit,
    onOpenApplications: () -> Unit,
    onSubmitted: () -> Unit,
    onPickLocation: () -> Unit,
    viewModel: RoomCreateViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    var typeSheetVisible by remember { mutableStateOf(false) }
    val currentOnSubmitted by rememberUpdatedState(onSubmitted)
    LaunchedEffect(viewModel) {
        viewModel.submitted.collect { currentOnSubmitted() }
    }
    val locationLabel = when {
        state.locating -> stringResource(R.string.location_locating)

        state.locationReady -> PlaceNames.display(
            state.placeName,
            isCurrentGps = true,
            nearbyLabel = stringResource(R.string.place_nearby),
        )

        else -> stringResource(R.string.room_create_center_hint)
    }
    val placeTypeLabel = state.placeType?.let { stringResource(RoomPlaceType.label(it)) }
        ?: stringResource(R.string.room_create_type_hint)

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.nearby_room_create)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    TextButton(
                        onClick = onOpenApplications,
                        colors = ButtonDefaults.textButtonColors(contentColor = AppColors.TextSecondary),
                    ) {
                        Text(stringResource(R.string.room_create_applications), fontSize = AppType.Secondary)
                    }
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .imePadding()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = AppSpacing.L, vertical = AppSpacing.S),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(AppShapes.CardLg)
                    .background(AppColors.CardBg),
            ) {
                RoomFormField(
                    label = stringResource(R.string.room_create_name),
                    value = state.name,
                    onValueChange = viewModel::onName,
                    placeholder = stringResource(R.string.room_create_name_hint),
                    singleLine = true,
                )
                RoomFormDivider()
                RoomFormField(
                    label = stringResource(R.string.room_create_desc),
                    value = state.description,
                    onValueChange = viewModel::onDesc,
                    placeholder = stringResource(R.string.room_create_desc_hint),
                    singleLine = false,
                    minLines = 2,
                )
                RoomFormDivider()
                RoomFormLinkRow(
                    label = stringResource(R.string.room_create_type),
                    value = placeTypeLabel,
                    placeholder = state.placeType == null,
                    onClick = { typeSheetVisible = true },
                )
                RoomFormDivider()
                RoomFormLinkRow(
                    label = stringResource(R.string.room_create_center),
                    value = locationLabel,
                    placeholder = !state.locationReady,
                    onClick = onPickLocation,
                    maxLines = 2,
                )
                RoomFormDivider()
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = AppSpacing.L, vertical = RoomFormRowVPad),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        stringResource(R.string.room_create_radius),
                        fontSize = AppType.Secondary,
                        color = AppColors.TextPrimary,
                        modifier = Modifier.width(RoomFormLabelWidth),
                    )
                    RoomRadiusSlider(
                        value = state.radius,
                        onValueChange = viewModel::onRadius,
                        valueRange = RoomCreateViewModel.RADIUS_MIN..RoomCreateViewModel.RADIUS_MAX,
                        modifier = Modifier.weight(1f),
                    )
                    Text(
                        stringResource(R.string.room_create_radius_value, state.radius.toInt()),
                        fontSize = AppType.Secondary,
                        color = AppColors.TextSecondary,
                        modifier = Modifier.padding(start = AppSpacing.M),
                    )
                }
            }

            AppPrimaryButton(
                text = stringResource(if (state.loading) R.string.room_create_submitting else R.string.room_create_submit),
                onClick = viewModel::create,
                enabled = !state.loading,
                modifier = Modifier.padding(top = AppSpacing.Xl),
            )

            Text(
                stringResource(R.string.room_create_footer),
                color = AppColors.TextTertiary,
                fontSize = AppType.Caption,
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 10.dp, bottom = AppSpacing.Xxl),
            )
        }
    }

    if (typeSheetVisible) {
        RoomPlaceTypeSheet(
            selectedType = state.placeType,
            onSelect = { type ->
                viewModel.onPlaceType(type)
                typeSheetVisible = false
            },
            onDismiss = { typeSheetVisible = false },
        )
    }
}
