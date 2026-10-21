export const MOBILE_PATTERN = /^1[3-9]\d{9}$/

export const LOGIN_COPY = {
  invalidMobile: '请输入正确的手机号',
  emptyCode: '请输入验证码',
  needAgreement: '请先阅读并同意《用户协议》与《隐私政策》',
  codeSent: '验证码已发送',
  sendingCode: '发送中…',
  getCode: '获取验证码',
  login: '登录',
  loggingIn: '登录中…',
} as const

export const LOGIN_CODE_COUNTDOWN_SEC = 60
