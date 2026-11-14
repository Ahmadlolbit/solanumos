return {
  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = function()
        if not pcall(vim.cmd.colorscheme, "solanum") then
          vim.cmd.colorscheme("tokyonight")
        end
      end,
    },
  },
}
