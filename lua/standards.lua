-- Praetor Neovim LSP and Tool Configuration
local lspconfig = require("lspconfig")
local configs = require("lspconfig.configs")

vim.api.nvim_create_user_command("StandardsVerifyAll", function()
  vim.cmd("!make verify-all")
end, { desc = "Run the repository command Standards: Verify All" })
