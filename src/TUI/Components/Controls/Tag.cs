using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;
using TUI.Engine.Theme;
using TUI.UserInterface;

namespace TUI.Components.Controls;

public class Tag : ComponentAttribute
{
    private IEnumerable<string> _tags;
    private string _gitType;

    public void Bind(IEnumerable<string> tags, string gitType)
    {
        _tags = tags;
        _gitType = gitType;
    }

    public void Render(Horizontal horizontal, Size size)
    {
        var tagBuilder = new StringBuilder();

        tagBuilder.Append(GetGitTypeImage(_gitType));
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(_tags.Have("public") ? Icons.NetworkPublic : Icons.NetworkPrivate);
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(_tags.Have("seo") ? Icons.SEO : Icons.SEO.Disable());
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(_tags.Have("auth") ? Icons.Auth : Icons.Auth.Disable());
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(GetApplicationType());
        tagBuilder.Append(Symbols.Space);

        // base.Render(tagBuilder, position, size);
    }

    private string GetApplicationType()
    {
        foreach (var application in Icons.Applications)
            if (_tags.Have(application.Value))
                return application.Key;

        return Icons.Undefined;
    }

    private static char GetGitTypeImage(string gitType) =>
        gitType switch
        {
            "gitlab" => Symbols.GitLab,
            "github" => Symbols.GitHub,
            _ => Symbols.Git
        };

    protected override Sketch DrawComponent()
    {
        throw new NotImplementedException();
    }
}