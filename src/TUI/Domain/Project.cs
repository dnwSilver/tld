using System.Runtime.CompilerServices;
using TUI.Settings;
using YamlDotNet.Serialization;


namespace TUI.Domain;

[YamlSerializable]
public class Project
{
    [YamlMember]
    public string Icon { get; set; }

    [YamlMember]
    public string Name { get; set; }

    [YamlMember]
    public DependencyDto[] Dependencies { get; set; }

    [YamlMember]
    public IList<SourceDto> Sources { get; set; }
}