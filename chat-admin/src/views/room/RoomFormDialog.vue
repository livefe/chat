<template>
  <el-dialog
    :model-value="visible"
    :title="form.id ? '编辑聊天区' : '新增聊天区'"
    width="960px"
    destroy-on-close
    :close-on-click-modal="false"
    @update:model-value="emit('update:visible', $event)"
  >
    <el-form ref="formRef" :model="form" :rules="roomFormRules" label-width="100px">
      <el-form-item label="名称" prop="name">
        <el-input v-model="name" placeholder="请输入名称" maxlength="64" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input
          v-model="description"
          type="textarea"
          :rows="2"
          placeholder="简单介绍这个聊天区"
          maxlength="200"
        />
      </el-form-item>
      <el-form-item label="类型" prop="placeType">
        <el-select v-model="placeType" placeholder="请选择场所类型" style="width: 240px">
          <el-option
            v-for="item in ROOM_PLACE_TYPE_OPTIONS"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="围栏限制">
        <el-switch v-model="requireGeofence" active-text="仅围栏内可聊天" inactive-text="不限制" />
        <div class="form-tip">默认不限制；开启后用户需位于地理围栏内方可查看消息和发言</div>
      </el-form-item>
      <el-form-item label="聊天区形状">
        <AmapPicker
          :shape-type="form.shapeType"
          :center-lat="form.centerLat"
          :center-lng="form.centerLng"
          :radius="form.radius"
          :polygon="form.polygon"
          @change="emit('shape-change', $event)"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:visible', false)">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onConfirm">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { FormInstance } from 'element-plus'
import AmapPicker from '@/components/amap/AmapPicker.vue'
import { roomFormRules, type RoomFormModel } from '@/composables/useRoomForm'
import { ROOM_PLACE_TYPE_OPTIONS } from '@/constants/room'

defineProps<{
  visible: boolean
  submitting: boolean
  form: RoomFormModel
}>()

const name = defineModel<string>('name', { required: true })
const description = defineModel<string>('description', { required: true })
const placeType = defineModel<number>('placeType')
const requireGeofence = defineModel<boolean>('requireGeofence', { required: true })

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
  'shape-change': [
    payload: {
      shapeType: 1 | 2
      centerLat?: number
      centerLng?: number
      radius?: number
      polygon?: [number, number][]
    },
  ]
}>()

const formRef = ref<FormInstance>()

async function onConfirm() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (valid) emit('submit')
}
</script>
