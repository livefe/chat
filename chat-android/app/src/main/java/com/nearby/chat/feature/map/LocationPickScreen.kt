package com.nearby.chat.feature.map

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
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
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.appTopAppBarColors

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LocationPickScreen(
    onBack: () -> Unit,
    onPicked: (LocationPickResult) -> Unit,
    viewModel: LocationPickViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val currentOnPicked by rememberUpdatedState(onPicked)
    var cityPickerVisible by remember { mutableStateOf(false) }

    LaunchedEffect(viewModel) {
        viewModel.picked.collect { currentOnPicked(it) }
    }

    Scaffold(
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = {
            TopAppBar(
                title = { AppTopBarTitle(stringResource(R.string.location_pick_title)) },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                windowInsets = WindowInsets.statusBars,
                colors = appTopAppBarColors(AppColors.Surface),
            )
        },
    ) { padding ->
        Box(
            Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            if (state.ready) {
                TiandituMapView(
                    initial = state.center,
                    moveTo = state.flyTo,
                    modifier = Modifier.fillMaxSize(),
                    onCenterChanged = viewModel::onCenter,
                )
            }
            LocationPickSearchPanel(
                state = state,
                onOpenCityPicker = { cityPickerVisible = true },
                onKeyword = viewModel::onKeyword,
                onSearch = viewModel::search,
                onClearKeyword = viewModel::clearKeyword,
                onSelectPoi = viewModel::selectPoi,
                modifier = Modifier.align(Alignment.TopCenter),
            )
            LocationPickBottomPanel(
                state = state,
                onUseCurrentLocation = viewModel::useCurrentLocation,
                onConfirm = viewModel::confirmPick,
                modifier = Modifier.align(Alignment.BottomCenter),
            )
        }
    }

    RegionCityPickerSheet(
        visible = cityPickerVisible,
        onDismiss = { cityPickerVisible = false },
        onSelect = { region ->
            cityPickerVisible = false
            viewModel.selectRegion(region)
        },
    )
}
