namespace TUI.Pages;

interface IPage
{
    void Open();
    void Initial();
    void Render();
    void Bind();
    void Load();
}