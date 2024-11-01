using System.Text;
using TUI.Domain;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Theme;
using TUI.UserInterface;
using static TUI.Engine.Symbols;

namespace TUI.Controls.Components;

public class ProjectTitle : ComponentBase
{
    private readonly Project _project;
    
    public ProjectTitle(Project project)
    {
        _project = project;
    }
    
    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();
        builder.Append(GetHub().Colorized());
        builder.Append(Space);
        builder.Append((_project.IsPublicNetwork ? NetworkPublic : NetworkPrivate).Colorized());
        builder.Append(Space);
        builder.Append(_project.SeoDependent ? Seo.Colorized() : Seo.Disable());
        builder.Append(Space);
        builder.Append(_project.HasAuth ? Auth.Colorized() : Auth.Disable());
        builder.Append(Space);
        builder.Append(GetApplicationType().Colorized());
        builder.Append(Space);
        builder.Append(_project.Name.Disable());
        return new Sketch(builder);
    }
    
    private string GetHub() => _project.Hub.Type == "gitlab" ? GitLab : GitHub;
    
    private string GetApplicationType()
    {
        foreach (var application in Icons.Applications.Where(application => _project.Tags.Have(application.Value)))
            return application.Key;
        
        return Undefined.Hint();
    }
}