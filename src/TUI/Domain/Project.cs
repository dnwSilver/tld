using System.Runtime.CompilerServices;
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
    public Dependency[] Dependencies { get; set; }

    [YamlMember]
    public IList<Source> Sources { get; set; }
}