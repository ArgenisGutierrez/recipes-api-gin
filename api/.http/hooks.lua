local show_result = require("http-nvim.hooks_utils").show
local update_env = require("http-nvim.hooks_utils").update_env

local function ask_for_confirmation(request, start_request)
  local confirmation = vim.fn.input("Are you sure you want to run this request? [y/N] ")

  if confirmation == "y" or confirmation == "Y" then
    start_request()
  end
end

local function save_access_token(request, response, stdout)
  show_result(request, response)

  if response.status_code ~= 200 then
    return
  end

  local body = vim.json.decode(response.body)

  update_env({
    access_token = body.access_token,
    refresh_token = body.refresh_token,
  })
end

return {
  save_access_token = save_access_token,
  ask_for_confirmation = ask_for_confirmation,
}
