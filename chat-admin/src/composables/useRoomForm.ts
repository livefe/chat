import { reactive, ref } from 'vue'
import { ElMessage, type FormRules } from 'element-plus'
import { createRoom, getRoomDetail, updateRoom, type RoomItem } from '@/api/room'
import { gcj02ToWgs84, hasCoordinates, wgs84ToGcj02 } from '@/utils/geo'
import { showApiError } from '@/utils/apiError'

export interface RoomFormModel {
  id: number
  name: string
  description: string
  placeType: number | undefined
  shapeType: 1 | 2
  centerLat: number
  centerLng: number
  radius: number
  polygon: [number, number][]
  requireGeofence: boolean
}

export const roomFormRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  placeType: [{ required: true, message: '请选择场所类型', trigger: 'change' }],
}

export function useRoomForm(onSaved: () => void) {
  const formVisible = ref(false)
  const submitting = ref(false)

  const form = reactive<RoomFormModel>({
    id: 0,
    name: '',
    description: '',
    placeType: undefined,
    shapeType: 1,
    centerLat: 0,
    centerLng: 0,
    radius: 500,
    polygon: [],
    requireGeofence: false,
  })

  function resetForm() {
    form.id = 0
    form.name = ''
    form.description = ''
    form.placeType = undefined
    form.shapeType = 1
    form.centerLat = 0
    form.centerLng = 0
    form.radius = 500
    form.polygon = []
    form.requireGeofence = false
  }

  function handleAdd() {
    resetForm()
    formVisible.value = true
  }

  async function handleEdit(row: RoomItem) {
    try {
      const res = await getRoomDetail(row.id)
      const d = res.data
      form.id = d.id
      form.name = d.name
      form.description = d.description
      form.placeType = d.placeType || undefined
      form.shapeType = d.shapeType as 1 | 2
      if (d.shapeType === 1 && hasCoordinates(d.centerLat, d.centerLng)) {
        const center = wgs84ToGcj02(d.centerLat, d.centerLng)
        form.centerLat = center.lat
        form.centerLng = center.lng
      } else {
        form.centerLat = d.centerLat
        form.centerLng = d.centerLng
      }
      form.radius = d.radius
      form.polygon = (d.polygon || []).map(([lng, lat]) => {
        const p = wgs84ToGcj02(lat, lng)
        return [p.lng, p.lat] as [number, number]
      })
      form.requireGeofence = !!d.requireGeofence
      formVisible.value = true
    } catch (e) {
      showApiError(e, '获取聊天区详情失败')
    }
  }

  function onShapeChange(payload: {
    shapeType: 1 | 2
    centerLat?: number
    centerLng?: number
    radius?: number
    polygon?: [number, number][]
  }) {
    form.shapeType = payload.shapeType
    if (payload.shapeType === 1) {
      form.centerLat = payload.centerLat ?? 0
      form.centerLng = payload.centerLng ?? 0
      form.radius = payload.radius ?? 500
      form.polygon = []
    } else {
      form.polygon = payload.polygon || []
    }
  }

  async function submit() {
    if (form.shapeType === 1) {
      if (!hasCoordinates(form.centerLat, form.centerLng)) {
        ElMessage.warning('请在地图上选择圆心位置')
        return
      }
      if (!form.radius) {
        ElMessage.warning('请设置半径')
        return
      }
    } else {
      if (!form.polygon.length) {
        ElMessage.warning('请在地图上绘制多边形')
        return
      }
    }
    submitting.value = true
    try {
      const submitCenter =
        form.shapeType === 1
          ? gcj02ToWgs84(form.centerLat, form.centerLng)
          : { lat: form.centerLat, lng: form.centerLng }
      const submitPolygon =
        form.shapeType === 2
          ? form.polygon.map(([lng, lat]) => {
              const p = gcj02ToWgs84(lat, lng)
              return [p.lng, p.lat] as [number, number]
            })
          : form.polygon
      const payload = {
        name: form.name,
        description: form.description,
        placeType: form.placeType as number,
        shapeType: form.shapeType,
        requireGeofence: form.requireGeofence,
        centerLat: form.shapeType === 1 ? submitCenter.lat : undefined,
        centerLng: form.shapeType === 1 ? submitCenter.lng : undefined,
        radius: form.shapeType === 1 ? form.radius : undefined,
        polygon: form.shapeType === 2 ? submitPolygon : undefined,
      }
      if (form.id) {
        await updateRoom({ id: form.id, ...payload })
        ElMessage.success('更新成功')
      } else {
        await createRoom(payload)
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      onSaved()
    } catch (e) {
      showApiError(e, '保存失败')
    } finally {
      submitting.value = false
    }
  }

  return {
    formVisible,
    submitting,
    form,
    resetForm,
    handleAdd,
    handleEdit,
    submit,
    onShapeChange,
  }
}
