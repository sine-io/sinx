<script setup lang="ts">
import type { PropType } from 'vue'

interface MenuNode {
  id: number | string
  name: string
  path?: string
  children?: MenuNode[]
}

defineOptions({ name: 'MenuTree' })

defineProps({
  nodes: {
    type: Array as PropType<MenuNode[]>,
    default: () => [],
  },
})

function keyOf(node: MenuNode) {
  return node.path || String(node.id)
}
</script>

<template>
  <template v-for="node in nodes" :key="keyOf(node)">
    <a-sub-menu v-if="node.children && node.children.length" :key="keyOf(node)" :title="node.name">
      <MenuTree :nodes="node.children" />
    </a-sub-menu>
    <a-menu-item v-else :key="keyOf(node)">{{ node.name }}</a-menu-item>
  </template>
</template>
