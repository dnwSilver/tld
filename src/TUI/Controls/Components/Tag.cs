using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;
using TUI.Engine.Theme;
using TUI.UserInterface;

namespace TUI.Controls.Components;

public class Tag : ComponentBase
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
        tagBuilder.Append(_tags.Have("public") ? Symbols.NetworkPublic : Symbols.NetworkPrivate);
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(_tags.Have("seo") ? Symbols.SEO : Symbols.SEO.Disable());
        tagBuilder.Append(Symbols.Space);
        tagBuilder.Append(_tags.Have("auth") ? Symbols.Auth : Symbols.Auth.Disable());
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

        return Symbols.Undefined;
    }

    private static string GetGitTypeImage(string gitType) =>
        gitType switch
        {
            "gitlab" => Symbols.GitLab,
            "github" => Symbols.GitHub,
            _ => Symbols.Git
        };

    protected override Sketch DrawComponent(Size minSize)
    {
        throw new NotImplementedException();
    }
}