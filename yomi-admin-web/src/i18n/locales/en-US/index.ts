import auth from './auth.json'
import common from './common.json'
import dashboard from './dashboard.json'
import errorPage from './errorPage.json'
import errors from './errors.json'
import layout from './layout.json'
import menu from './menu.json'
import httpRecord from './system/httpRecord.json'
import param from './system/param.json'
import permission from './system/permission.json'
import permissionGroup from './system/permissionGroup.json'
import role from './system/role.json'
import user from './system/user.json'
import validation from './validation.json'

export default {
  common,
  menu,
  auth,
  layout,
  dashboard,
  validation,
  errors,
  errorPage,
  system: {
    user,
    role,
    permissionGroup,
    permission,
    httpRecord,
    param,
  },
}
