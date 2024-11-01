using System.Runtime.Serialization;
using YamlDotNet.Serialization;

namespace TUI.Providers.Dependencies;

[YamlSerializable]
public class ProjectDto
{
    [DataMember] [YamlMember] public int Id { get; set; }

    [DataMember] [YamlMember] public string Name { get; set; }

    [DataMember] [YamlMember] public string Deps { get; set; }

    [DataMember] [YamlMember] public IEnumerable<string> Tags { get; set; }
}