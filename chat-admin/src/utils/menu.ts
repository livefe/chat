import { type PermissionItem } from '@/api/permission'
import { compareBySortThenId, routeNameFromCode } from '@/utils/permission'
import { PermissionType } from '@/constants/permission'

export interface MenuItem {
  path: string
  name: string
  title: string
  icon?: string
  permission?: string
  children?: MenuItem[]
}

export interface HomeEntry {
  title: string
  path: string
  icon?: string
  permission?: string
}

export interface HomeSection {
  title: string
  items: HomeEntry[]
}

type MenuNode = Omit<MenuItem, 'children'> & {
  children: MenuNode[]
  __parentId: number
  __sort: number
  __id: number
}

function sortMenuNodes(nodes: MenuNode[]): MenuItem[] {
  return nodes
    .slice()
    .sort((a, b) =>
      compareBySortThenId({ sort: a.__sort, id: a.__id }, { sort: b.__sort, id: b.__id }),
    )
    .map(({ __parentId, __sort, __id, children, ...item }) => ({
      ...item,
      children: children.length ? sortMenuNodes(children) : undefined,
    }))
}

export function buildMenus(permissions: PermissionItem[]): MenuItem[] {
  const menus = permissions.filter(
    (p) => p.type === PermissionType.Menu && p.status === 1 && !p.hideInMenu,
  )

  const map = new Map<number, MenuNode>()
  const tree: MenuNode[] = []

  menus.forEach((p) => {
    map.set(p.id, {
      path: p.path,
      name: routeNameFromCode(p.code),
      title: p.name,
      icon: p.icon,
      permission: p.code,
      children: [],
      __parentId: p.parentId,
      __sort: p.sort,
      __id: p.id,
    })
  })

  map.forEach((node) => {
    if (node.__parentId === 0) {
      tree.push(node)
    } else {
      const parent = map.get(node.__parentId)
      if (parent) {
        parent.children.push(node)
      } else {
        tree.push(node)
      }
    }
  })

  return sortMenuNodes(tree)
}

export function buildHomeSections(permissions: PermissionItem[]): HomeSection[] {
  const tree = buildMenus(permissions)
  const sections: HomeSection[] = []

  for (const node of tree) {
    const leaves = (node.children || []).filter((c) => !!c.path)
    if (leaves.length > 0) {
      sections.push({
        title: node.title,
        items: leaves.map((c) => ({
          title: c.title,
          path: c.path,
          icon: c.icon || node.icon,
          permission: c.permission,
        })),
      })
      continue
    }
    if (node.path && node.path !== '/home') {
      sections.push({
        title: node.title,
        items: [
          {
            title: node.title,
            path: node.path,
            icon: node.icon,
            permission: node.permission,
          },
        ],
      })
    }
  }

  return sections.filter((s) => s.items.length > 0)
}

export function filterMenus(menus: MenuItem[], keyword: string): MenuItem[] {
  const kw = keyword.trim().toLowerCase()
  if (!kw) return menus
  const out: MenuItem[] = []
  for (const m of menus) {
    if (m.title.toLowerCase().includes(kw)) {
      out.push(m)
      continue
    }
    const children = m.children ? filterMenus(m.children, kw) : []
    if (children.length) out.push({ ...m, children })
  }
  return out
}

export function flattenMenuLeaves(menus: MenuItem[]): MenuItem[] {
  const out: MenuItem[] = []
  for (const m of menus) {
    if (m.children?.length) out.push(...flattenMenuLeaves(m.children))
    else if (m.path) out.push(m)
  }
  return out
}

export function findMenuTrail(menus: MenuItem[], path: string): MenuItem[] {
  function walk(nodes: MenuItem[], trail: MenuItem[]): MenuItem[] | null {
    for (const node of nodes) {
      const next = [...trail, node]
      if (node.path === path) return next
      if (node.children?.length) {
        const hit = walk(node.children, next)
        if (hit) return hit
      }
    }
    return null
  }
  return walk(menus, []) || []
}
